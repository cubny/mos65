package cpu

import (
	"fmt"
	"math/rand"

	"github.com/cubny/mos65/internal/memory"
)

// Status register flag bits.
const (
	FlagC byte = 0x01
	FlagZ byte = 0x02
	FlagI byte = 0x04
	FlagD byte = 0x08
	FlagB byte = 0x10
	FlagU byte = 0x20 // unused, conventionally 1
	FlagV byte = 0x40
	FlagN byte = 0x80
)

// CPU is the 6502 simulator state.
type CPU struct {
	A, X, Y byte
	P       byte
	PC      uint16
	SP      byte

	Mem     *memory.Memory
	Running bool

	// Optional callbacks.
	OnMessage func(string)
	Rand      func() byte
}

// New returns a CPU bound to mem with default registers.
func New(mem *memory.Memory) *CPU {
	c := &CPU{Mem: mem}
	c.Reset()
	return c
}

// Reset zeroes registers and program memory (0x0000-0x05ff) and sets PC to the
// program origin.
func (c *CPU) Reset() {
	c.Mem.Clear(0, memory.ProgramStart)
	c.A, c.X, c.Y = 0, 0, 0
	c.PC = memory.ProgramStart
	c.SP = 0xff
	c.P = FlagB | FlagU
	c.Running = false
}

func (c *CPU) message(s string) {
	if c.OnMessage != nil {
		c.OnMessage(s)
	}
}

func (c *CPU) randomByte() byte {
	if c.Rand != nil {
		return c.Rand()
	}
	return byte(rand.Intn(256))
}

func (c *CPU) popByte() byte {
	v := c.Mem.Get(c.PC)
	c.PC++
	return v
}

func (c *CPU) popWord() uint16 {
	lo := uint16(c.popByte())
	hi := uint16(c.popByte())
	return lo | hi<<8
}

func (c *CPU) stackPush(v byte) {
	c.Mem.Set(0x0100+uint16(c.SP), v)
	if c.SP == 0 {
		c.SP = 0xff
		c.message("6502 Stack filled! Wrapping...")
	} else {
		c.SP--
	}
}

func (c *CPU) stackPop() byte {
	if c.SP == 0xff {
		c.SP = 0
		c.message("6502 Stack emptied! Wrapping...")
	} else {
		c.SP++
	}
	return c.Mem.Get(0x0100 + uint16(c.SP))
}

// setNVflags updates the N and Z flags based on the (8-bit) result.
func (c *CPU) setNVflags(v byte) {
	if v == 0 {
		c.P |= FlagZ
	} else {
		c.P &^= FlagZ
	}
	if v&0x80 != 0 {
		c.P |= FlagN
	} else {
		c.P &^= FlagN
	}
}

func (c *CPU) setCarryFromBit0(v byte) {
	c.P = (c.P & 0xfe) | (v & 1)
}

func (c *CPU) setCarryFromBit7(v byte) {
	c.P = (c.P & 0xfe) | ((v >> 7) & 1)
}

func (c *CPU) carrySet() bool    { return c.P&FlagC != 0 }
func (c *CPU) zeroSet() bool     { return c.P&FlagZ != 0 }
func (c *CPU) negativeSet() bool { return c.P&FlagN != 0 }
func (c *CPU) overflowSet() bool { return c.P&FlagV != 0 }
func (c *CPU) decimalMode() bool { return c.P&FlagD != 0 }

func (c *CPU) clc()         { c.P &^= FlagC }
func (c *CPU) sec()         { c.P |= FlagC }
func (c *CPU) clv()         { c.P &^= FlagV }
func (c *CPU) setOverflow() { c.P |= FlagV }

func (c *CPU) jumpBranch(offset byte) {
	if offset > 0x7f {
		c.PC = c.PC - (0x100 - uint16(offset))
	} else {
		c.PC = c.PC + uint16(offset)
	}
}

func (c *CPU) doCompare(reg, val byte) {
	if reg >= val {
		c.sec()
	} else {
		c.clc()
	}
	c.setNVflags(reg - val)
}

// Step executes one instruction. Returns false when the program should halt
// (either BRK set Running=false or PC wrapped back to 0).
func (c *CPU) Step() {
	c.Mem.Set(memory.RandomAddr, c.randomByte())
	opcode := c.popByte()
	c.dispatch(opcode)
}

// StatusFlags returns the status register as an 8-character "NV-BDIZC" string.
func (c *CPU) StatusFlags() string {
	out := make([]byte, 8)
	for i := 0; i < 8; i++ {
		if c.P>>(7-i)&1 == 1 {
			out[i] = '1'
		} else {
			out[i] = '0'
		}
	}
	return string(out)
}

// DebugLine returns a short multi-line snapshot of CPU state.
func (c *CPU) DebugLine() string {
	return fmt.Sprintf("A=$%02x X=$%02x Y=$%02x\nSP=$%02x PC=$%04x\nNV-BDIZC\n%s",
		c.A, c.X, c.Y, c.SP, c.PC, c.StatusFlags())
}
