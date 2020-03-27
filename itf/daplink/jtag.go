//-----------------------------------------------------------------------------
/*

CMSIS-DAP JTAG Driver

*/
//-----------------------------------------------------------------------------

package daplink

import (
	"errors"
	"fmt"
	"time"

	"github.com/deadsy/hidapi"
	"github.com/deadsy/rvdbg/bitstr"
	"github.com/deadsy/rvdbg/jtag"
)

//-----------------------------------------------------------------------------

// jtagSeq is a JTAG sequence element.
type jtagSeq struct {
	info byte
	tdi  []byte
}

const infoBits = (63 << 0)
const infoTms = (1 << 6)
const infoTdo = (1 << 7)

// nBits returns the number of bits for a JTAG sequence element.
func (s *jtagSeq) nBits() int {
	n := int(s.info & infoBits)
	if n == 0 {
		n = 64
	}
	return n
}

// nTdiBytes returns the number of TDI bytes for a JTAG sequence element.
func (s *jtagSeq) nTdiBytes() int {
	return (s.nBits() + 7) >> 3
}

// nTdoBytes returns the number of TDO bytes for a JTAG sequence element.
func (s *jtagSeq) nTdoBytes() int {
	n := 0
	if s.info&infoTdo != 0 {
		n = s.nBits()
	}
	return (n + 7) >> 3
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// bitStringToJtagSeq converts a bit string to a JTAG sequence.
func bitStringToJtagSeq(bs *bitstr.BitString, needTdo bool) []jtagSeq {

	data := bs.GetBytes()
	n := bs.Len()
	seq := []jtagSeq{}
	idx := 0

	for n > 0 {
		k := min(n, 64)
		end := min(idx+8, len(data))
		info := byte(k & infoBits)
		if needTdo {
			info |= infoTdo
		}
		seq = append(seq, jtagSeq{info, data[idx:end]})
		idx += 8
		n -= k
	}

	return seq
}

//-----------------------------------------------------------------------------

// Jtag is a driver for CMSIS-DAP JTAG operations.
type Jtag struct {
	dev *device
}

func (j *Jtag) String() string {
	return fmt.Sprintf("%s", j.dev)
}

// NewJtag returns a new CMSIS-DAP JTAG driver.
func NewJtag(devInfo *hidapi.DeviceInfo, speed int) (*Jtag, error) {

	// get the hid device
	hid, err := hidapi.Open(devInfo.VendorID, devInfo.ProductID, devInfo.SerialNumber)
	if err != nil {
		return nil, err
	}

	dev, err := newDevice(hid)
	if err != nil {
		hid.Close()
		return nil, err
	}

	j := &Jtag{
		dev: dev,
	}

	// make sure the CMSIS-DAP can do JTAG
	if !j.dev.hasCap(capJtag) {
		j.Close()
		return nil, errors.New("jtag not supported")
	}

	// connect in JTAG mode
	err = j.dev.cmdConnect(modeJtag)
	if err != nil {
		j.Close()
		return nil, err
	}

	// set the clock speed
	err = j.dev.cmdSwjClock(speed)
	if err != nil {
		j.Close()
		return nil, err
	}

	return j, nil
}

// Close closes a CMSIS-DAP JTAG driver.
func (j *Jtag) Close() error {
	j.dev.cmdDisconnect()
	j.dev.close()
	return nil
}

// GetState returns the JTAG hardware state.
func (j *Jtag) GetState() (*jtag.State, error) {
	pins, err := j.dev.getPins()
	if err != nil {
		return nil, err
	}
	return &jtag.State{
		Tck:  pins&pinTCK != 0,
		Tdi:  pins&pinTDI != 0,
		Tdo:  pins&pinTDO != 0,
		Tms:  pins&pinTMS != 0,
		Trst: pins&pinTRST != 0,
		Srst: pins&pinSRST != 0,
	}, nil
}

// HasCapability returns true if this driver has the indicated capability.
func (j *Jtag) HasCapability(capability jtag.Capability) bool {
	switch capability {
	}
	return false
}

// TestReset pulses the test reset line.
func (j *Jtag) TestReset(delay time.Duration) error {
	err := j.dev.setPins(pinTRST)
	if err != nil {
		return err
	}
	time.Sleep(delay)
	return j.dev.clrPins(pinTRST)
}

// SystemReset pulses the system reset line.
func (j *Jtag) SystemReset(delay time.Duration) error {
	err := j.dev.setPins(pinSRST)
	if err != nil {
		return err
	}
	time.Sleep(delay)
	return j.dev.clrPins(pinSRST)
}

// TapReset resets the TAP state machine.
func (j *Jtag) TapReset() error {
	return j.dev.cmdSwjSequence(jtag.ToIdle)
}

// scanXR handles the back half of an IR/DR scan ooperation
func (j *Jtag) scanXR(tdi *bitstr.BitString, idle uint, needTdo bool) (*bitstr.BitString, error) {
	rx, err := j.dev.cmdJtagSequence(bitStringToJtagSeq(tdi, needTdo))
	if err != nil {
		return nil, err
	}
	err = j.dev.cmdSwjSequence(jtag.ShiftToIdle[idle])
	if err != nil {
		return nil, err
	}
	if !needTdo {
		return nil, nil
	}
	return bitstr.FromBytes(rx, tdi.Len()), nil
}

// ScanIR scans bits through the JTAG IR chain
func (j *Jtag) ScanIR(tdi *bitstr.BitString, needTdo bool) (*bitstr.BitString, error) {
	err := j.dev.cmdSwjSequence(jtag.IdleToIRshift)
	if err != nil {
		return nil, err
	}
	return j.scanXR(tdi, 0, needTdo)
}

// ScanDR scans bits through the JTAG DR chain
func (j *Jtag) ScanDR(tdi *bitstr.BitString, idle uint, needTdo bool) (*bitstr.BitString, error) {
	err := j.dev.cmdSwjSequence(jtag.IdleToDRshift)
	if err != nil {
		return nil, err
	}
	return j.scanXR(tdi, idle, needTdo)
}

//-----------------------------------------------------------------------------