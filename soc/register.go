//-----------------------------------------------------------------------------
/*

Peripheral Registers

*/
//-----------------------------------------------------------------------------

package soc

import (
	"fmt"
	"strings"
)

//-----------------------------------------------------------------------------

// Register is peripheral register.
type Register struct {
	Name   string
	Offset uint
	Size   uint
	Descr  string
	Fields []Field
	parent *Peripheral
}

// RegisterSet is a set of registers.
type RegisterSet []Register

//-----------------------------------------------------------------------------
// Sort registers by offset.

func (a RegisterSet) Len() int      { return len(a) }
func (a RegisterSet) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a RegisterSet) Less(i, j int) bool {
	// Offsets for registers may not be unique.
	// Tie break with the name to give a well-defined sort order.
	if a[i].Offset == a[j].Offset {
		return strings.Compare(a[i].Name, a[j].Name) < 0
	}
	return a[i].Offset < a[j].Offset
}

//-----------------------------------------------------------------------------

// address returns the absolute address of an indexed register.
func (r *Register) address(base, idx uint) uint {
	return base + r.Offset + (idx * (r.Size >> 3))
}

// Register returns strings for the decode of a register.
func (r *Register) Display(drv Driver, fields bool) [][]string {

	addr := r.address(r.parent.Addr, 0)
	val := uint(0) // TODO

	// address string
	fmtStr := fmt.Sprintf(": %s[%%d:0]", getAddressFormat(drv))
	addrStr := fmt.Sprintf(fmtStr, addr, r.Size-1)

	// has the value changed?
	changed := "  "

	// value string
	var valStr string
	if val == 0 {
		valStr = fmt.Sprintf("= 0%s", changed)
	} else {
		fmtStr := fmt.Sprintf("= 0x%%0%dx%%s", r.Size>>2)
		valStr = fmt.Sprintf(fmtStr, val, changed)
	}

	s := [][]string{}
	s = append(s, []string{r.Name, addrStr, valStr, r.Descr})
	// field decodes
	if fields && len(r.Fields) != 0 {
		for i := range r.Fields {
			s = append(s, r.Fields[i].Display(val))
		}
	}

	return s
}

//-----------------------------------------------------------------------------