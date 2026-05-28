package assembler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cubny/mos65/internal/memory"
)

// Origin is the address where code is assembled by default.
const Origin = memory.ProgramStart

// Error describes a single-line assembly failure.
type Error struct {
	Line             int    // 1-based line number
	Source           string // the offending source line, sanitized
	OutOfRangeBranch bool
}

func (e *Error) Error() string {
	if e.OutOfRangeBranch {
		return fmt.Sprintf("out of range branch on line %d (branches are limited to -128 to +127): %s",
			e.Line, e.Source)
	}
	return fmt.Sprintf("syntax error on line %d: %s", e.Line, e.Source)
}

// Result describes a successful assembly.
type Result struct {
	CodeLen int
	EndPC   uint16
}

// Assemble parses src and writes machine code starting at Origin into mem.
// Memory 0x0000..0x05ff is zeroed first to match the JS simulator behavior.
func Assemble(mem *memory.Memory, src string) (Result, error) {
	mem.Clear(0, memory.ProgramStart)

	lines := strings.Split(src+"\n\n", "\n")

	a := &assembler{mem: mem, labels: map[string]uint16{}}
	a.pc = Origin

	// Preprocess: strip comments, collect `define NAME value` directives, blank
	// directive lines in-place.
	a.symbols = preprocess(lines)

	// Pass 1: index labels by walking lines and measuring how many bytes each
	// emits. Forward references emit filler that pass 2 will overwrite.
	a.pc = Origin
	for i, line := range lines {
		a.codeLen = 0
		startPC := a.pc
		a.assembleLine(line)
		// label after the assembleLine call so labels resolve to the line's
		// starting PC, mirroring the JS index pass.
		if m := reLabelDef.FindStringSubmatch(line); m != nil {
			label := m[1]
			if _, ok := a.symbols[label]; ok {
				return Result{}, &Error{Line: i + 1, Source: line}
			}
			if _, ok := a.labels[label]; ok {
				return Result{}, &Error{Line: i + 1, Source: line}
			}
			a.labels[label] = startPC
		}
	}

	// Pass 2: real emit with all labels known.
	a.pc = Origin
	a.codeLen = 0
	for i, line := range lines {
		a.wasOutOfRangeBranch = false
		if !a.assembleLine(line) {
			return Result{}, &Error{
				Line:             i + 1,
				Source:           line,
				OutOfRangeBranch: a.wasOutOfRangeBranch,
			}
		}
	}

	if a.codeLen == 0 {
		return Result{}, fmt.Errorf("no code to run")
	}
	mem.Set(a.pc, 0x00) // null byte terminator
	return Result{CodeLen: a.codeLen, EndPC: a.pc}, nil
}

type assembler struct {
	mem                 *memory.Memory
	pc                  uint16
	codeLen             int
	labels              map[string]uint16
	symbols             map[string]string
	wasOutOfRangeBranch bool
}

func (a *assembler) pushByte(v byte) {
	a.mem.Set(a.pc, v)
	a.pc++
	a.codeLen++
}

func (a *assembler) pushWord(v uint16) {
	a.pushByte(byte(v & 0xff))
	a.pushByte(byte(v >> 8))
}

func (a *assembler) lookupSymbol(name string) (string, bool) {
	v, ok := a.symbols[name]
	return v, ok
}

func (a *assembler) findLabel(name string) (uint16, bool) {
	v, ok := a.labels[name]
	return v, ok
}

var (
	reLabelDef     = regexp.MustCompile(`^(\w+):`)
	reLabelAndCmd  = regexp.MustCompile(`^\w+:\s*(\w+.*)$`)
	reCmdOnly      = regexp.MustCompile(`^(\w+)`)
	reOrigin       = regexp.MustCompile(`^\*\s*=\s*\$?[0-9a-fA-F]*$`)
	reWordParam    = regexp.MustCompile(`^\w+\s+(.*)$`)
	reBareCmd      = regexp.MustCompile(`^\w+$`)
	reImmediate    = regexp.MustCompile(`^#([\w\$%]+)$`)
	reImmLabelHilo = regexp.MustCompile(`^#([<>])(\w+)$`)
	reIndirect     = regexp.MustCompile(`^\(([\w\$]+)\)$`)
	reIndirectX    = regexp.MustCompile(`(?i)^\(([\w\$]+),X\)$`)
	reIndirectY    = regexp.MustCompile(`(?i)^\(([\w\$]+)\),Y$`)
	reIndexX       = regexp.MustCompile(`(?i)^([\w\$]+),X$`)
	reIndexY       = regexp.MustCompile(`(?i)^([\w\$]+),Y$`)
	reBare         = regexp.MustCompile(`^([\w\$]+)$`)
	reWord         = regexp.MustCompile(`^\w+$`)
	reDecByte      = regexp.MustCompile(`^([0-9]{1,3})$`)
	reHexByte      = regexp.MustCompile(`^\$([0-9a-fA-F]{1,2})$`)
	reBinByte      = regexp.MustCompile(`^%([01]{1,8})$`)
	reHexWord      = regexp.MustCompile(`^\$([0-9a-fA-F]{3,4})$`)
	reDecWord      = regexp.MustCompile(`^([0-9]{1,5})$`)
)

// assembleLine parses and emits bytes for a single (already-sanitized) line.
// It returns false on syntax error or out-of-range branch.
func (a *assembler) assembleLine(input string) bool {
	if input == "" {
		return true
	}

	// Strip an optional "label:" prefix.
	if reLabelDef.MatchString(input) {
		if m := reLabelAndCmd.FindStringSubmatch(input); m != nil {
			input = m[1]
		} else {
			return true
		}
	}
	if input == "" {
		return true
	}

	// Origin directive (`*=$0800` or `*=2048`).
	if reOrigin.MatchString(input) {
		raw := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(input), "*"))
		raw = strings.TrimSpace(strings.TrimPrefix(raw, "="))
		var addr int64
		var err error
		if strings.HasPrefix(raw, "$") {
			addr, err = strconv.ParseInt(raw[1:], 16, 32)
		} else {
			addr, err = strconv.ParseInt(raw, 10, 32)
		}
		if err != nil || addr < 0 || addr > 0xffff {
			return false
		}
		a.pc = uint16(addr)
		return true
	}

	cmdMatch := reCmdOnly.FindStringSubmatch(input)
	if cmdMatch == nil {
		return false
	}
	command := strings.ToUpper(cmdMatch[1])

	// Extract the parameter (everything after the mnemonic), then strip spaces.
	var param string
	if m := reWordParam.FindStringSubmatch(input); m != nil {
		param = strings.ReplaceAll(m[1], " ", "")
	} else if reBareCmd.MatchString(input) {
		param = ""
	} else {
		return false
	}

	if command == "DCB" {
		return a.dcb(param)
	}

	entry := findMnemonic(command)
	if entry == nil {
		return false
	}

	// Order matters: SNGL first to handle accumulator form, then ZP/X/Y modes
	// before ABS (to prefer narrower encoding), then IND* variants, then ABS,
	// then BRA.
	if a.checkSingle(param, entry.Modes[ModeSNGL]) {
		return true
	}
	if a.checkImmediate(param, entry.Modes[ModeImm]) {
		return true
	}
	if a.checkZeroPage(param, entry.Modes[ModeZP]) {
		return true
	}
	if a.checkZeroPageX(param, entry.Modes[ModeZPX]) {
		return true
	}
	if a.checkZeroPageY(param, entry.Modes[ModeZPY]) {
		return true
	}
	if a.checkAbsoluteX(param, entry.Modes[ModeABSX]) {
		return true
	}
	if a.checkAbsoluteY(param, entry.Modes[ModeABSY]) {
		return true
	}
	if a.checkIndirect(param, entry.Modes[ModeIND]) {
		return true
	}
	if a.checkIndirectX(param, entry.Modes[ModeINDX]) {
		return true
	}
	if a.checkIndirectY(param, entry.Modes[ModeINDY]) {
		return true
	}
	if a.checkAbsolute(param, entry.Modes[ModeABS]) {
		return true
	}
	if a.checkBranch(param, entry.Modes[ModeBRA]) {
		return true
	}
	return false
}

func (a *assembler) dcb(param string) bool {
	if param == "" {
		return false
	}
	values := strings.Split(param, ",")
	for _, raw := range values {
		if raw == "" {
			continue
		}
		var n int64
		var err error
		switch raw[0] {
		case '$':
			n, err = strconv.ParseInt(raw[1:], 16, 32)
		case '%':
			n, err = strconv.ParseInt(raw[1:], 2, 32)
		default:
			if raw[0] >= '0' && raw[0] <= '9' {
				n, err = strconv.ParseInt(raw, 10, 32)
			} else {
				return false
			}
		}
		if err != nil {
			return false
		}
		a.pushByte(byte(n & 0xff))
	}
	return true
}

func (a *assembler) tryParseByte(param string) (byte, bool) {
	if reWord.MatchString(param) {
		if v, ok := a.lookupSymbol(param); ok {
			param = v
		}
	}
	if m := reDecByte.FindStringSubmatch(param); m != nil {
		n, _ := strconv.ParseInt(m[1], 10, 32)
		if n >= 0 && n <= 0xff {
			return byte(n), true
		}
	}
	if m := reHexByte.FindStringSubmatch(param); m != nil {
		n, _ := strconv.ParseInt(m[1], 16, 32)
		return byte(n), true
	}
	if m := reBinByte.FindStringSubmatch(param); m != nil {
		n, _ := strconv.ParseInt(m[1], 2, 32)
		return byte(n), true
	}
	return 0, false
}

func (a *assembler) tryParseWord(param string) (uint16, bool) {
	if reWord.MatchString(param) {
		if v, ok := a.lookupSymbol(param); ok {
			param = v
		}
	}
	if m := reHexWord.FindStringSubmatch(param); m != nil {
		n, _ := strconv.ParseInt(m[1], 16, 32)
		if n >= 0 && n <= 0xffff {
			return uint16(n), true
		}
	}
	if m := reDecWord.FindStringSubmatch(param); m != nil {
		n, _ := strconv.ParseInt(m[1], 10, 32)
		if n >= 0 && n <= 0xffff {
			return uint16(n), true
		}
	}
	return 0, false
}

func (a *assembler) checkSingle(param string, opcode int16) bool {
	if opcode == none {
		return false
	}
	if param != "" && param != "A" && param != "a" {
		return false
	}
	a.pushByte(byte(opcode))
	return true
}

func (a *assembler) checkImmediate(param string, opcode int16) bool {
	if opcode == none {
		return false
	}
	if m := reImmediate.FindStringSubmatch(param); m != nil {
		if v, ok := a.tryParseByte(m[1]); ok {
			a.pushByte(byte(opcode))
			a.pushByte(v)
			return true
		}
	}
	if m := reImmLabelHilo.FindStringSubmatch(param); m != nil {
		hilo, label := m[1], m[2]
		a.pushByte(byte(opcode))
		if addr, ok := a.findLabel(label); ok {
			switch hilo {
			case ">":
				a.pushByte(byte(addr >> 8))
				return true
			case "<":
				a.pushByte(byte(addr & 0xff))
				return true
			}
			return false
		}
		a.pushByte(0)
		return true
	}
	return false
}

func (a *assembler) checkZeroPage(param string, opcode int16) bool {
	if opcode == none {
		return false
	}
	if v, ok := a.tryParseByte(param); ok {
		a.pushByte(byte(opcode))
		a.pushByte(v)
		return true
	}
	return false
}

func (a *assembler) checkZeroPageX(param string, opcode int16) bool {
	if opcode == none {
		return false
	}
	if m := reIndexX.FindStringSubmatch(param); m != nil {
		if v, ok := a.tryParseByte(m[1]); ok {
			a.pushByte(byte(opcode))
			a.pushByte(v)
			return true
		}
	}
	return false
}

func (a *assembler) checkZeroPageY(param string, opcode int16) bool {
	if opcode == none {
		return false
	}
	if m := reIndexY.FindStringSubmatch(param); m != nil {
		if v, ok := a.tryParseByte(m[1]); ok {
			a.pushByte(byte(opcode))
			a.pushByte(v)
			return true
		}
	}
	return false
}

func (a *assembler) checkAbsolute(param string, opcode int16) bool {
	if opcode == none {
		return false
	}
	if m := reBare.FindStringSubmatch(param); m != nil {
		if v, ok := a.tryParseWord(m[1]); ok {
			a.pushByte(byte(opcode))
			a.pushWord(v)
			return true
		}
	}
	if reWord.MatchString(param) {
		a.pushByte(byte(opcode))
		if addr, ok := a.findLabel(param); ok {
			a.pushWord(addr)
		} else {
			a.pushWord(0xffff)
		}
		return true
	}
	return false
}

func (a *assembler) checkAbsoluteX(param string, opcode int16) bool {
	if opcode == none {
		return false
	}
	m := reIndexX.FindStringSubmatch(param)
	if m == nil {
		return false
	}
	if v, ok := a.tryParseWord(m[1]); ok {
		a.pushByte(byte(opcode))
		a.pushWord(v)
		return true
	}
	if !reWord.MatchString(m[1]) {
		return false
	}
	a.pushByte(byte(opcode))
	if addr, ok := a.findLabel(m[1]); ok {
		a.pushWord(addr)
	} else {
		a.pushWord(0xffff)
	}
	return true
}

func (a *assembler) checkAbsoluteY(param string, opcode int16) bool {
	if opcode == none {
		return false
	}
	m := reIndexY.FindStringSubmatch(param)
	if m == nil {
		return false
	}
	if v, ok := a.tryParseWord(m[1]); ok {
		a.pushByte(byte(opcode))
		a.pushWord(v)
		return true
	}
	if !reWord.MatchString(m[1]) {
		return false
	}
	a.pushByte(byte(opcode))
	if addr, ok := a.findLabel(m[1]); ok {
		a.pushWord(addr)
	} else {
		a.pushWord(0xffff)
	}
	return true
}

func (a *assembler) checkIndirect(param string, opcode int16) bool {
	if opcode == none {
		return false
	}
	if m := reIndirect.FindStringSubmatch(param); m != nil {
		if v, ok := a.tryParseWord(m[1]); ok {
			a.pushByte(byte(opcode))
			a.pushWord(v)
			return true
		}
	}
	return false
}

func (a *assembler) checkIndirectX(param string, opcode int16) bool {
	if opcode == none {
		return false
	}
	if m := reIndirectX.FindStringSubmatch(param); m != nil {
		if v, ok := a.tryParseByte(m[1]); ok {
			a.pushByte(byte(opcode))
			a.pushByte(v)
			return true
		}
	}
	return false
}

func (a *assembler) checkIndirectY(param string, opcode int16) bool {
	if opcode == none {
		return false
	}
	if m := reIndirectY.FindStringSubmatch(param); m != nil {
		if v, ok := a.tryParseByte(m[1]); ok {
			a.pushByte(byte(opcode))
			a.pushByte(v)
			return true
		}
	}
	return false
}

func (a *assembler) checkBranch(param string, opcode int16) bool {
	if opcode == none {
		return false
	}
	addr, ok := a.findLabel(param)
	if !ok {
		// Reserve two bytes during pass 1 so PC stays aligned with pass 2.
		a.pushWord(0)
		return false
	}
	a.pushByte(byte(opcode))
	distance := int(addr) - int(a.pc) - 1
	if distance < -128 || distance > 127 {
		a.wasOutOfRangeBranch = true
		return false
	}
	a.pushByte(byte(distance))
	return true
}
