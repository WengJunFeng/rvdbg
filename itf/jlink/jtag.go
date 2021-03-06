//-----------------------------------------------------------------------------
/*

Segger J-Link JTAG Driver

*/
//-----------------------------------------------------------------------------

package jlink

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/deadsy/jaylink"
	"github.com/deadsy/rvdbg/bitstr"
	"github.com/deadsy/rvdbg/jtag"
	"github.com/deadsy/rvdbg/util/log"
)

//-----------------------------------------------------------------------------

// Jtag is a driver for J-link JTAG operations.
type Jtag struct {
	dev     *jaylink.Device
	hdl     *jaylink.DeviceHandle
	version jaylink.JtagVersion
	speed   int // current JTAG clock speed in kHz
}

func (j *Jtag) String() string {
	s := []string{}
	hw, err := j.hdl.GetHardwareVersion()
	if err == nil {
		s = append(s, fmt.Sprintf("hardware %s", hw))
	}
	sn, err := j.dev.GetSerialNumber()
	if err == nil {
		s = append(s, fmt.Sprintf("serial number %d", sn))
	}
	ver, err := j.hdl.GetFirmwareVersion()
	if err == nil {
		s = append(s, fmt.Sprintf("firmware %s", ver))
	}
	state, err := j.hdl.GetHardwareStatus()
	if err == nil {
		s = append(s, fmt.Sprintf("target voltage %dmV", state.TargetVoltage))
	}
	s = append(s, fmt.Sprintf("jtag speed %dkHz", j.speed))
	return strings.Join(s, "\n")
}

// NewJtag returns a new J-Link JTAG driver.
func NewJtag(dev *jaylink.Device, speed int) (*Jtag, error) {
	j := &Jtag{
		dev: dev,
	}

	// get the device handle
	hdl, err := dev.Open()
	if err != nil {
		return nil, err
	}
	j.hdl = hdl

	// get the device capabilities
	caps, err := hdl.GetAllCaps()
	if err != nil {
		hdl.Close()
		return nil, err
	}

	// get the JTAG command version
	version, err := hdl.GetJtagCommandVersion()
	if err != nil {
		hdl.Close()
		return nil, err
	}
	j.version = version

	// check and select the target interface
	if caps.HasCap(jaylink.DEV_CAP_SELECT_TIF) {
		itf, err := hdl.GetAvailableInterfaces()
		if err != nil {
			hdl.Close()
			return nil, err
		}
		if itf&(1<<jaylink.TIF_JTAG) == 0 {
			hdl.Close()
			return nil, errors.New("jtag interface not available")
		}
		_, err = hdl.SelectInterface(jaylink.TIF_JTAG)
		if err != nil {
			hdl.Close()
			return nil, err
		}
	} else {
		// Target interface selection is not supported. Assume JTAG is auto-selected.
		log.Info.Printf("DEV_CAP_SELECT_TIF not supported, assuming JTAG is auto-selected\n")
	}

	// check the desired interface speed
	if caps.HasCap(jaylink.DEV_CAP_GET_SPEEDS) {
		maxSpeed, err := hdl.GetMaxSpeed()
		if err != nil {
			hdl.Close()
			return nil, err
		}
		if speed > int(maxSpeed) {
			log.Info.Printf("JTAG speed %dkHz is too high, limiting to %dkHz (max)", speed, maxSpeed)
			speed = int(maxSpeed)
		}
	}

	// set the interface speed
	err = hdl.SetSpeed(uint16(speed))
	if err != nil {
		hdl.Close()
		return nil, err
	}
	j.speed = speed

	return j, nil
}

// Close closes a J-Link JTAG driver.
func (j *Jtag) Close() error {
	return j.hdl.Close()
}

// GetState returns the JTAG hardware state.
func (j *Jtag) GetState() (*jtag.State, error) {
	status, err := j.hdl.GetHardwareStatus()
	if err != nil {
		return nil, err
	}
	return &jtag.State{
		TargetVoltage: int(status.TargetVoltage),
		Tck:           status.Tck,
		Tdi:           status.Tdi,
		Tdo:           status.Tdo,
		Tms:           status.Tms,
		Trst:          status.Trst,
		Srst:          status.Tres,
	}, nil
}

// jtagIO performs jtag IO operations.
func (j *Jtag) jtagIO(tms, tdi *bitstr.BitString, needTdo bool) (*bitstr.BitString, error) {
	tdo, err := j.hdl.JtagIO(tms.GetBytes(), tdi.GetBytes(), uint16(tdi.Len()), j.version)
	if needTdo {
		return bitstr.FromBytes(tdo, tdi.Len()), err
	}
	return nil, err
}

// TestReset pulses the test reset line.
func (j *Jtag) TestReset(delay time.Duration) error {
	err := j.hdl.JtagClearTrst()
	if err != nil {
		return err
	}
	time.Sleep(delay)
	return j.hdl.JtagSetTrst()
}

// SystemReset pulses the system reset line.
func (j *Jtag) SystemReset(delay time.Duration) error {
	err := j.hdl.ClearReset()
	if err != nil {
		return err
	}
	time.Sleep(delay)
	return j.hdl.SetReset()
}

// TapReset resets the TAP state machine.
func (j *Jtag) TapReset() error {
	tdi := bitstr.Zeros(jtag.ToIdle.Len())
	_, err := j.jtagIO(jtag.ToIdle, tdi, false)
	return err
}

// ScanIR scans bits through the JTAG IR chain
func (j *Jtag) ScanIR(tdi *bitstr.BitString, needTdo bool) (*bitstr.BitString, error) {
	shiftToIdle := jtag.ShiftToIdle[0]
	tms := bitstr.Null().Tail(jtag.IdleToIRshift).Tail0(tdi.Len() - 1).Tail(shiftToIdle)
	tdi = bitstr.Zeros(jtag.IdleToIRshift.Len()).Tail(tdi).Tail0(shiftToIdle.Len() - 1)
	tdo, err := j.jtagIO(tms, tdi, needTdo)
	if err != nil {
		return nil, err
	}
	if needTdo {
		tdo.DropHead(jtag.IdleToIRshift.Len()).DropTail(shiftToIdle.Len() - 1)
		return tdo, nil
	}
	return nil, nil
}

// ScanDR scans bits through the JTAG DR chain
func (j *Jtag) ScanDR(tdi *bitstr.BitString, idle uint, needTdo bool) (*bitstr.BitString, error) {
	shiftToIdle := jtag.ShiftToIdle[idle]
	tms := bitstr.Null().Tail(jtag.IdleToDRshift).Tail0(tdi.Len() - 1).Tail(shiftToIdle)
	tdi = bitstr.Zeros(jtag.IdleToDRshift.Len()).Tail(tdi).Tail0(shiftToIdle.Len() - 1)
	tdo, err := j.jtagIO(tms, tdi, needTdo)
	if err != nil {
		return nil, err
	}
	if needTdo {
		tdo.DropHead(jtag.IdleToDRshift.Len()).DropTail(shiftToIdle.Len() - 1)
		return tdo, nil
	}
	return nil, nil
}

//-----------------------------------------------------------------------------
