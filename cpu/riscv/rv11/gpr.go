//-----------------------------------------------------------------------------
/*

RISC-V Debugger 0.11
general purpose register access

*/
//-----------------------------------------------------------------------------

package rv11

import (
	"fmt"

	"github.com/deadsy/rvdbg/cpu/riscv/rv"
)

//-----------------------------------------------------------------------------

// rdGPR reads a GPR using debug ram operations.
func rdGPR(dbg *Debug, reg, size uint) (uint64, error) {

	if size == 32 {
		dbg.cache.wr(0, rv.InsSW(reg, ramAddr(0), rv.RegZero))
		dbg.cache.wrResume(1)
		dbg.cache.read(0)
		// run the code
		err := dbg.cache.flush(true)
		if err != nil {
			return 0, err
		}
		return uint64(dbg.cache.rd(0)), nil
	}

	if size == 64 {
		dbg.cache.wr(0, rv.InsSD(reg, ramAddr(2), rv.RegZero))
		dbg.cache.wrResume(1)
		dbg.cache.read(2)
		dbg.cache.read(3)
		// run the code
		err := dbg.cache.flush(true)
		if err != nil {
			return 0, err
		}
		lx := uint64(dbg.cache.rd(2))
		ux := uint64(dbg.cache.rd(3))
		return (ux << 32) | lx, nil
	}

	return 0, fmt.Errorf("%d-bit gpr reads are not supported", size)
}

// wrGPR writes a GPR using debug ram operations.
func wrGPR(dbg *Debug, reg, size uint, val uint64) error {

	if size == 32 {
		dbg.cache.wr(0, rv.InsLW(reg, ramAddr(2), rv.RegZero))
		dbg.cache.wrResume(1)
		dbg.cache.wr(2, uint32(val))
		// run the code
		return dbg.cache.flush(true)
	}

	if size == 64 {
		dbg.cache.wr(0, rv.InsLD(reg, ramAddr(2), rv.RegZero))
		dbg.cache.wrResume(1)
		dbg.cache.wr(2, uint32(val))
		dbg.cache.wr(3, uint32(val>>32))
		// run the code
		return dbg.cache.flush(true)
	}

	return fmt.Errorf("%d-bit gpr writes are not supported", size)
}

//-----------------------------------------------------------------------------
