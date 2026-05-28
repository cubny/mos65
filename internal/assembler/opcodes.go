package assembler

// Addressing-mode column indices, matching the order of fields in opEntry.Modes.
const (
	ModeImm  = 0
	ModeZP   = 1
	ModeZPX  = 2
	ModeZPY  = 3
	ModeABS  = 4
	ModeABSX = 5
	ModeABSY = 6
	ModeIND  = 7
	ModeINDX = 8
	ModeINDY = 9
	ModeSNGL = 10
	ModeBRA  = 11
)

// none is the sentinel for "this mode is not valid for this mnemonic".
const none int16 = -1

type opEntry struct {
	Name  string
	Modes [12]int16
}

// opcodes mirrors the JS Opcodes table column-for-column. -1 means "not valid".
var opcodes = []opEntry{
	{"ADC", [12]int16{0x69, 0x65, 0x75, none, 0x6d, 0x7d, 0x79, none, 0x61, 0x71, none, none}},
	{"AND", [12]int16{0x29, 0x25, 0x35, none, 0x2d, 0x3d, 0x39, none, 0x21, 0x31, none, none}},
	{"ASL", [12]int16{none, 0x06, 0x16, none, 0x0e, 0x1e, none, none, none, none, 0x0a, none}},
	{"BIT", [12]int16{none, 0x24, none, none, 0x2c, none, none, none, none, none, none, none}},
	{"BPL", [12]int16{none, none, none, none, none, none, none, none, none, none, none, 0x10}},
	{"BMI", [12]int16{none, none, none, none, none, none, none, none, none, none, none, 0x30}},
	{"BVC", [12]int16{none, none, none, none, none, none, none, none, none, none, none, 0x50}},
	{"BVS", [12]int16{none, none, none, none, none, none, none, none, none, none, none, 0x70}},
	{"BCC", [12]int16{none, none, none, none, none, none, none, none, none, none, none, 0x90}},
	{"BCS", [12]int16{none, none, none, none, none, none, none, none, none, none, none, 0xb0}},
	{"BNE", [12]int16{none, none, none, none, none, none, none, none, none, none, none, 0xd0}},
	{"BEQ", [12]int16{none, none, none, none, none, none, none, none, none, none, none, 0xf0}},
	{"BRK", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x00, none}},
	{"CMP", [12]int16{0xc9, 0xc5, 0xd5, none, 0xcd, 0xdd, 0xd9, none, 0xc1, 0xd1, none, none}},
	{"CPX", [12]int16{0xe0, 0xe4, none, none, 0xec, none, none, none, none, none, none, none}},
	{"CPY", [12]int16{0xc0, 0xc4, none, none, 0xcc, none, none, none, none, none, none, none}},
	{"DEC", [12]int16{none, 0xc6, 0xd6, none, 0xce, 0xde, none, none, none, none, none, none}},
	{"EOR", [12]int16{0x49, 0x45, 0x55, none, 0x4d, 0x5d, 0x59, none, 0x41, 0x51, none, none}},
	{"CLC", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x18, none}},
	{"SEC", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x38, none}},
	{"CLI", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x58, none}},
	{"SEI", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x78, none}},
	{"CLV", [12]int16{none, none, none, none, none, none, none, none, none, none, 0xb8, none}},
	{"CLD", [12]int16{none, none, none, none, none, none, none, none, none, none, 0xd8, none}},
	{"SED", [12]int16{none, none, none, none, none, none, none, none, none, none, 0xf8, none}},
	{"INC", [12]int16{none, 0xe6, 0xf6, none, 0xee, 0xfe, none, none, none, none, none, none}},
	{"JMP", [12]int16{none, none, none, none, 0x4c, none, none, 0x6c, none, none, none, none}},
	{"JSR", [12]int16{none, none, none, none, 0x20, none, none, none, none, none, none, none}},
	{"LDA", [12]int16{0xa9, 0xa5, 0xb5, none, 0xad, 0xbd, 0xb9, none, 0xa1, 0xb1, none, none}},
	{"LDX", [12]int16{0xa2, 0xa6, none, 0xb6, 0xae, none, 0xbe, none, none, none, none, none}},
	{"LDY", [12]int16{0xa0, 0xa4, 0xb4, none, 0xac, 0xbc, none, none, none, none, none, none}},
	{"LSR", [12]int16{none, 0x46, 0x56, none, 0x4e, 0x5e, none, none, none, none, 0x4a, none}},
	{"NOP", [12]int16{none, none, none, none, none, none, none, none, none, none, 0xea, none}},
	{"ORA", [12]int16{0x09, 0x05, 0x15, none, 0x0d, 0x1d, 0x19, none, 0x01, 0x11, none, none}},
	{"TAX", [12]int16{none, none, none, none, none, none, none, none, none, none, 0xaa, none}},
	{"TXA", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x8a, none}},
	{"DEX", [12]int16{none, none, none, none, none, none, none, none, none, none, 0xca, none}},
	{"INX", [12]int16{none, none, none, none, none, none, none, none, none, none, 0xe8, none}},
	{"TAY", [12]int16{none, none, none, none, none, none, none, none, none, none, 0xa8, none}},
	{"TYA", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x98, none}},
	{"DEY", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x88, none}},
	{"INY", [12]int16{none, none, none, none, none, none, none, none, none, none, 0xc8, none}},
	{"ROR", [12]int16{none, 0x66, 0x76, none, 0x6e, 0x7e, none, none, none, none, 0x6a, none}},
	{"ROL", [12]int16{none, 0x26, 0x36, none, 0x2e, 0x3e, none, none, none, none, 0x2a, none}},
	{"RTI", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x40, none}},
	{"RTS", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x60, none}},
	{"SBC", [12]int16{0xe9, 0xe5, 0xf5, none, 0xed, 0xfd, 0xf9, none, 0xe1, 0xf1, none, none}},
	{"STA", [12]int16{none, 0x85, 0x95, none, 0x8d, 0x9d, 0x99, none, 0x81, 0x91, none, none}},
	{"TXS", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x9a, none}},
	{"TSX", [12]int16{none, none, none, none, none, none, none, none, none, none, 0xba, none}},
	{"PHA", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x48, none}},
	{"PLA", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x68, none}},
	{"PHP", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x08, none}},
	{"PLP", [12]int16{none, none, none, none, none, none, none, none, none, none, 0x28, none}},
	{"STX", [12]int16{none, 0x86, none, 0x96, 0x8e, none, none, none, none, none, none, none}},
	{"STY", [12]int16{none, 0x84, 0x94, none, 0x8c, none, none, none, none, none, none, none}},
	{"WDM", [12]int16{0x42, 0x42, none, none, none, none, none, none, none, none, none, none}},
}

// findMnemonic returns the opEntry for a mnemonic (uppercased), or nil.
func findMnemonic(name string) *opEntry {
	for i := range opcodes {
		if opcodes[i].Name == name {
			return &opcodes[i]
		}
	}
	return nil
}
