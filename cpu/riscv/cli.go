//-----------------------------------------------------------------------------
/*

RISC-V CPU Menu Items

*/
//-----------------------------------------------------------------------------

package riscv

import (
	"fmt"
	"math"
	"strings"

	cli "github.com/deadsy/go-cli"
	"github.com/deadsy/rvdbg/cpu/riscv/rv"
	"github.com/deadsy/rvdbg/soc"
)

//-----------------------------------------------------------------------------

// target provides a method for getting the CPU debugger driver.
type target interface {
	GetRiscvDebug() rv.Debug
	GetCSR() (*soc.Device, soc.Driver)
}

//-----------------------------------------------------------------------------
// display CSR

// csrHelp is help information for the "csr" command.
var csrHelp = []cli.Help{
	{"[register]", "register (string) - register name (or *)"},
	{"<cr>", "display all registers"},
}

// cmdCSR displays the control and status registers.
var cmdCSR = cli.Leaf{
	Descr: "display control and status registers",
	F: func(c *cli.CLI, args []string) {

		err := cli.CheckArgc(args, []int{0, 1})
		if err != nil {
			c.User.Put(fmt.Sprintf("%s\n", err))
			return
		}

		csr, drv := c.User.(target).GetCSR()

		p := csr.GetPeripheral("CSR")

		if len(args) == 0 {
			c.User.Put(fmt.Sprintf("%s\n", p.Display(drv, nil, false)))
			return
		}

		if args[0] == "*" {
			c.User.Put(fmt.Sprintf("%s\n", p.Display(drv, nil, true)))
			return
		}

		r := p.GetRegister(args[0])
		if r == nil {
			c.User.Put(fmt.Sprintf("no register \"%s\" (run \"csr\" for the names)\n", args[0]))
			return
		}

		c.User.Put(fmt.Sprintf("%s\n", p.Display(drv, r, true)))
	},
}

//-----------------------------------------------------------------------------
// display general purpose register set

var abiXName = [32]string{
	"zero", "ra", "sp", "gp", "tp", "t0", "t1", "t2",
	"s0", "s1", "a0", "a1", "a2", "a3", "a4", "a5",
	"a6", "a7", "s2", "s3", "s4", "s5", "s6", "s7",
	"s8", "s9", "s10", "s11", "t3", "t4", "t5", "t6",
}

var regCache []uint64

func gprString(reg []uint64, xlen uint) string {
	fmtx := "%08x"
	if xlen == 64 {
		fmtx = "%016x"
	}
	if regCache == nil {
		regCache = reg
	}
	s := make([]string, len(reg))
	for i := 0; i < len(reg); i++ {
		delta := ""
		if reg[i] != regCache[i] {
			delta = " *"
		}
		if i == len(reg)-1 {
			s[i] = fmt.Sprintf("%-9s "+fmtx+"%s", "pc", reg[i], delta)
		} else {
			regStr := fmt.Sprintf("x%d", i)
			valStr := "0"
			if reg[i] != 0 {
				valStr = fmt.Sprintf(fmtx, reg[i])
			}
			s[i] = fmt.Sprintf("%-4s %-4s %s%s", regStr, abiXName[i], valStr, delta)
		}
	}
	regCache = reg
	return strings.Join(s, "\n")
}

// CmdGpr displays the general purpose registers.
var CmdGpr = cli.Leaf{
	Descr: "display general purpose registers",
	F: func(c *cli.CLI, args []string) {
		dbg := c.User.(target).GetRiscvDebug()
		hi := dbg.GetCurrentHart()
		err := dbg.HaltHart()
		if err != nil {
			c.User.Put(fmt.Sprintf("unable to halt hart%d: %v\n", hi.ID, err))
			return
		}
		// slice of register values, +1 for the pc
		reg := make([]uint64, hi.Nregs+1)
		// read the GPRs
		for i := 0; i < hi.Nregs; i++ {
			var err error
			reg[i], err = dbg.RdGPR(uint(i), 0)
			if err != nil {
				c.User.Put(fmt.Sprintf("unable to read gpr%d: %v\n", i, err))
				return
			}
		}
		// read the PC
		pc, err := dbg.RdCSR(rv.DPC, 0)
		if err != nil {
			c.User.Put(fmt.Sprintf("unable to read pc: %v\n", err))
			return
		}
		reg[len(reg)-1] = pc
		c.User.Put(fmt.Sprintf("%s\n", gprString(reg, hi.MXLEN)))
	},
}

//-----------------------------------------------------------------------------
// display floating point register set

var abiFName = [32]string{
	"ft0", "ft1", "ft2", "ft3", "ft4", "ft5", "ft6", "ft7",
	"fs0", "fs1", "fa0", "fa1", "fa2", "fa3", "fa4", "fa5",
	"fa6", "fa7", "fs2", "fs3", "fs4", "fs5", "fs6", "fs7",
	"fs8", "fs9", "fs10", "fs11", "ft8", "ft9", "ft10", "ft11",
}

func fprString(reg []uint64, flen int) string {
	s := make([]string, len(reg))
	for i := 0; i < len(reg); i++ {
		regStr := fmt.Sprintf("f%d", i)
		valStr := "0"
		if reg[i] != 0 {
			valStr = fmt.Sprintf("%016x", reg[i])
		}
		f32 := math.Float32frombits(uint32(reg[i]))
		s[i] = fmt.Sprintf("%-4s %-4s %-16s %f", regStr, abiFName[i], valStr, f32)
	}
	return strings.Join(s, "\n")
}

//-----------------------------------------------------------------------------

// CmdHalt halts the current hart.
var CmdHalt = cli.Leaf{
	Descr: "halt the current hart",
	F: func(c *cli.CLI, args []string) {
		dbg := c.User.(target).GetRiscvDebug()
		hi := dbg.GetCurrentHart()
		if hi.State == rv.Halted {
			c.User.Put(fmt.Sprintf("hart%d already halted\n", hi.ID))
			return
		}
		err := dbg.HaltHart()
		if err != nil {
			c.User.Put(fmt.Sprintf("unable to halt hart%d: %v\n", hi.ID, err))
			return
		}
	},
}

// CmdResume resumes the current hart.
var CmdResume = cli.Leaf{
	Descr: "resume the current hart",
	F: func(c *cli.CLI, args []string) {
		dbg := c.User.(target).GetRiscvDebug()
		hi := dbg.GetCurrentHart()
		if hi.State == rv.Running {
			c.User.Put(fmt.Sprintf("hart%d already running\n", hi.ID))
			return
		}
		err := dbg.ResumeHart()
		if err != nil {
			c.User.Put(fmt.Sprintf("unable to resume hart%d: %v\n", hi.ID, err))
			return
		}
	},
}

//-----------------------------------------------------------------------------

// HartHelp is help for the hart command.
var HartHelp = []cli.Help{
	{"<cr>", "display info for current hart"},
	{"<id>", "select hart<id> as the current hart"},
}

var CmdHart = cli.Leaf{
	Descr: "hart info/select",
	F: func(c *cli.CLI, args []string) {
		dbg := c.User.(target).GetRiscvDebug()
		hi := dbg.GetCurrentHart()
		if len(args) == 0 {
			c.User.Put(fmt.Sprintf("%s\n", hi))
			return
		}
		// TODO
	},
}

//-----------------------------------------------------------------------------

var cmdDebugInfo = cli.Leaf{
	Descr: "debug module information",
	F: func(c *cli.CLI, args []string) {
		dbg := c.User.(target).GetRiscvDebug()
		c.User.Put(fmt.Sprintf("%s\n", dbg.GetInfo()))
	},
}

var cmdRiscvTest1 = cli.Leaf{
	Descr: "test routine",
	F: func(c *cli.CLI, args []string) {
		dbg := c.User.(target).GetRiscvDebug()
		c.User.Put(fmt.Sprintf("%s\n", dbg.Test1()))
	},
}

var cmdRiscvTest2 = cli.Leaf{
	Descr: "test routine",
	F: func(c *cli.CLI, args []string) {
		dbg := c.User.(target).GetRiscvDebug()
		c.User.Put(fmt.Sprintf("%s\n", dbg.Test2()))
	},
}

// Menu submenu items
var Menu = cli.Menu{
	{"csr", cmdCSR, csrHelp},
	{"dmi", cmdDebugInfo},
	{"test1", cmdRiscvTest1},
	{"test2", cmdRiscvTest2},
}

//-----------------------------------------------------------------------------
