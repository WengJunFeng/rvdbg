//-----------------------------------------------------------------------------
/*

RISC-V Debugger API

*/
//-----------------------------------------------------------------------------

package rv

import (
	"fmt"

	cli "github.com/deadsy/go-cli"
	"github.com/deadsy/rvda"
	"github.com/deadsy/rvdbg/soc"
)

//-----------------------------------------------------------------------------

// HartState is the running state of a hart.
type HartState int

// HartState values.
const (
	Unknown HartState = iota // unknown
	Running                  // hart is running
	Halted                   // hart is halted
)

var stateName = map[HartState]string{
	Running: "running",
	Halted:  "halted",
}

func (s HartState) String() string {
	if name, ok := stateName[s]; ok {
		return name
	}
	return "unknown"
}

// HartInfo stores generic hart information.
type HartInfo struct {
	ID      int         // hart identifier
	State   HartState   // hart state
	Nregs   int         // number of GPRs (normally 32, 16 for rv32e)
	MXLEN   uint        // machine XLEN
	SXLEN   uint        // supervisor XLEN (0 == no S-mode)
	UXLEN   uint        // user XLEN (0 == no U-mode)
	HXLEN   uint        // hypervisor XLEN (0 == no H-mode)
	DXLEN   uint        // debug XLEN
	FLEN    uint        // foating point register width (0 == no floating point)
	MISA    uint        // MISA value
	MHARTID uint        // MHARTID value
	CSR     *soc.Device // CSR registers/fields
	ISA     *rvda.ISA   // ISA for the disassembler
}

func xlenString(n uint, msg string) string {
	if n != 0 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("no %s", msg)
}

func (hi *HartInfo) String() string {
	s := make([][]string, 0)
	s = append(s, []string{fmt.Sprintf("hart%d", hi.ID), fmt.Sprintf("%s", hi.State)})
	s = append(s, []string{"mhartid", fmt.Sprintf("%d", hi.MHARTID)})
	s = append(s, []string{"nregs", fmt.Sprintf("%d", hi.Nregs)})
	s = append(s, []string{"mxlen", fmt.Sprintf("%d", hi.MXLEN)})
	s = append(s, []string{"sxlen", xlenString(hi.SXLEN, "s-mode")})
	s = append(s, []string{"uxlen", xlenString(hi.UXLEN, "u-mode")})
	s = append(s, []string{"hxlen", xlenString(hi.HXLEN, "h-mode")})
	s = append(s, []string{"flen", xlenString(hi.FLEN, "floating point")})
	s = append(s, []string{"dxlen", fmt.Sprintf("%d", hi.DXLEN)})
	return cli.TableString(s, []int{0, 0}, 1)
}

//-----------------------------------------------------------------------------

// Debug is the RISC-V debug interface.
type Debug interface {
	GetPrompt(name string) string // get the target prompt
	// hart control
	GetHartCount() int                        // how many harts for this chip?
	GetHartInfo(id int) (*HartInfo, error)    // return the info structure for hart id
	GetCurrentHart() *HartInfo                // get the info structure for the current hart
	SetCurrentHart(id int) (*HartInfo, error) // set the current hart
	HaltHart() error                          // halt the current hart
	ResumeHart() error                        // resume the current hart
	// registers
	RdGPR(reg, size uint) (uint64, error)   // read general purpose register
	RdFPR(reg, size uint) (uint64, error)   // read floating point register
	RdCSR(reg, size uint) (uint64, error)   // read control and status register
	WrGPR(reg, size uint, val uint64) error // write general purpose register
	WrFPR(reg, size uint, val uint64) error // write floating point register
	WrCSR(reg, size uint, val uint64) error // write control and status register
	// memory
	GetAddressSize() uint                      // get address size in bits
	RdMem(width, addr, n uint) ([]uint, error) // read width-bit memory buffer
	WrMem(width, addr uint, val []uint) error  // write width-bit memory buffer
	// test
	Test1() string
	Test2() string
}

//-----------------------------------------------------------------------------
