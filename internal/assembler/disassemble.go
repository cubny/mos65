package assembler

import (
	"fmt"
	"strings"

	"github.com/cubny/mos65/internal/memory"
)

var modeNames = [...]string{
	ModeImm:  "Imm",
	ModeZP:   "ZP",
	ModeZPX:  "ZPX",
	ModeZPY:  "ZPY",
	ModeABS:  "ABS",
	ModeABSX: "ABSX",
	ModeABSY: "ABSY",
	ModeIND:  "IND",
	ModeINDX: "INDX",
	ModeINDY: "INDY",
	ModeSNGL: "SNGL",
	ModeBRA:  "BRA",
}

var instructionLength = map[string]int{
	"Imm":  2,
	"ZP":   2,
	"ZPX":  2,
	"ZPY":  2,
	"ABS":  3,
	"ABSX": 3,
	"ABSY": 3,
	"IND":  3,
	"INDX": 2,
	"INDY": 2,
	"SNGL": 1,
	"BRA":  2,
}

// lookupOpcode returns the mnemonic and mode for a given opcode byte. For
// unknown bytes it returns "???" / SNGL so the disassembler keeps moving.
func lookupOpcode(b byte) (name, mode string) {
	for i := range opcodes {
		for col, code := range opcodes[i].Modes {
			if code != none && byte(code) == b {
				return opcodes[i].Name, modeNames[col]
			}
		}
	}
	return "???", "SNGL"
}

func isAccumulatorOpcode(b byte) bool {
	switch b {
	case 0x0a, 0x2a, 0x4a, 0x6a:
		return true
	}
	return false
}

func isBranchMnemonic(name string) bool {
	if len(name) == 0 || name[0] != 'B' {
		return false
	}
	return name != "BIT" && name != "BRK"
}

// Disassemble walks codeLen bytes starting at Origin in mem and returns a
// formatted listing: "Address  Hexdump   Disassembly" header followed by one
// line per instruction.
func Disassemble(mem *memory.Memory, codeLen int) string {
	var b strings.Builder
	b.WriteString("Address  Hexdump   Disassembly\n")
	b.WriteString("-------------------------------\n")

	start := uint16(Origin)
	end := start + uint16(codeLen)
	for addr := start; addr < end; {
		op := mem.Get(addr)
		name, mode := lookupOpcode(op)
		length := instructionLength[mode]

		// Read operand bytes.
		operandBytes := make([]byte, 0, 2)
		for i := 1; i < length && addr+uint16(i) < end; i++ {
			operandBytes = append(operandBytes, mem.Get(addr+uint16(i)))
		}

		hexParts := []string{fmt.Sprintf("%02x", op)}
		for _, ob := range operandBytes {
			hexParts = append(hexParts, fmt.Sprintf("%02x", ob))
		}
		hexStr := strings.Join(hexParts, " ")
		padding := strings.Repeat(" ", 10-len(hexStr))
		if padding == "" {
			padding = " "
		}

		args := formatDisassemblyArgs(addr, op, name, mode, operandBytes)
		fmt.Fprintf(&b, "$%04x    %s%s%s %s\n", addr, hexStr, padding, name, args)

		addr += uint16(length)
	}
	return b.String()
}

func formatDisassemblyArgs(addr uint16, op byte, name, mode string, operands []byte) string {
	if isAccumulatorOpcode(op) {
		return "A"
	}

	// Operand bytes are little-endian; reverse for display.
	var argsString string
	if isBranchMnemonic(name) && len(operands) > 0 {
		off := operands[0]
		dest := int(addr) + 2
		if off > 0x7f {
			dest -= 0x100 - int(off)
		} else {
			dest += int(off)
		}
		argsString = fmt.Sprintf("%04x", uint16(dest))
	} else {
		for i := len(operands) - 1; i >= 0; i-- {
			argsString += fmt.Sprintf("%02x", operands[i])
		}
	}

	if argsString != "" {
		argsString = "$" + argsString
	}
	if mode == "Imm" {
		argsString = "#" + argsString
	}
	if strings.HasSuffix(mode, "X") {
		argsString += ",X"
	}
	if strings.HasPrefix(mode, "IND") {
		argsString = "(" + argsString + ")"
	}
	if strings.HasSuffix(mode, "Y") {
		argsString += ",Y"
	}
	return argsString
}

// Hexdump returns the standard 16-bytes-per-row hex dump of code memory.
func Hexdump(mem *memory.Memory, codeLen int) string {
	return mem.Format(Origin, codeLen)
}
