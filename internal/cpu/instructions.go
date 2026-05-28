package cpu

// dispatch executes one instruction given its opcode byte. The opcode comment
// after each case mirrors the names in the JS reference (i00 ... ife).
func (c *CPU) dispatch(op byte) {
	switch op {
	case 0x00: // BRK
		c.Running = false

	case 0x01: // ORA (zp,X)
		zp := byte(c.popByte() + c.X)
		addr := c.Mem.GetWord(uint16(zp))
		c.A |= c.Mem.Get(addr)
		c.setNVflags(c.A)

	case 0x05: // ORA zp
		c.A |= c.Mem.Get(uint16(c.popByte()))
		c.setNVflags(c.A)

	case 0x06: // ASL zp
		c.shiftLeftMem(uint16(c.popByte()))

	case 0x08: // PHP
		c.stackPush(c.P | 0x30)

	case 0x09: // ORA #imm
		c.A |= c.popByte()
		c.setNVflags(c.A)

	case 0x0a: // ASL A
		c.setCarryFromBit7(c.A)
		c.A = c.A << 1
		c.setNVflags(c.A)

	case 0x0d: // ORA abs
		c.A |= c.Mem.Get(c.popWord())
		c.setNVflags(c.A)

	case 0x0e: // ASL abs
		c.shiftLeftMem(c.popWord())

	case 0x10: // BPL
		off := c.popByte()
		if !c.negativeSet() {
			c.jumpBranch(off)
		}

	case 0x11: // ORA (zp),Y
		zp := c.popByte()
		addr := c.Mem.GetWord(uint16(zp)) + uint16(c.Y)
		c.A |= c.Mem.Get(addr)
		c.setNVflags(c.A)

	case 0x15: // ORA zp,X
		addr := byte(c.popByte() + c.X)
		c.A |= c.Mem.Get(uint16(addr))
		c.setNVflags(c.A)

	case 0x16: // ASL zp,X
		c.shiftLeftMem(uint16(byte(c.popByte() + c.X)))

	case 0x18: // CLC
		c.clc()

	case 0x19: // ORA abs,Y
		addr := c.popWord() + uint16(c.Y)
		c.A |= c.Mem.Get(addr)
		c.setNVflags(c.A)

	case 0x1d: // ORA abs,X
		addr := c.popWord() + uint16(c.X)
		c.A |= c.Mem.Get(addr)
		c.setNVflags(c.A)

	case 0x1e: // ASL abs,X
		c.shiftLeftMem(c.popWord() + uint16(c.X))

	case 0x20: // JSR
		addr := c.popWord()
		ret := c.PC - 1
		c.stackPush(byte(ret >> 8))
		c.stackPush(byte(ret & 0xff))
		c.PC = addr

	case 0x21: // AND (zp,X)
		zp := byte(c.popByte() + c.X)
		addr := c.Mem.GetWord(uint16(zp))
		c.A &= c.Mem.Get(addr)
		c.setNVflags(c.A)

	case 0x24: // BIT zp
		c.bit(c.Mem.Get(uint16(c.popByte())))

	case 0x25: // AND zp
		c.A &= c.Mem.Get(uint16(c.popByte()))
		c.setNVflags(c.A)

	case 0x26: // ROL zp
		c.rolMem(uint16(c.popByte()))

	case 0x28: // PLP
		c.P = c.stackPop() | 0x30 // B/U always set

	case 0x29: // AND #imm
		c.A &= c.popByte()
		c.setNVflags(c.A)

	case 0x2a: // ROL A
		sf := c.P & FlagC
		c.setCarryFromBit7(c.A)
		c.A = (c.A << 1) | sf
		c.setNVflags(c.A)

	case 0x2c: // BIT abs
		c.bit(c.Mem.Get(c.popWord()))

	case 0x2d: // AND abs
		c.A &= c.Mem.Get(c.popWord())
		c.setNVflags(c.A)

	case 0x2e: // ROL abs
		c.rolMem(c.popWord())

	case 0x30: // BMI
		off := c.popByte()
		if c.negativeSet() {
			c.jumpBranch(off)
		}

	case 0x31: // AND (zp),Y
		zp := c.popByte()
		addr := c.Mem.GetWord(uint16(zp)) + uint16(c.Y)
		c.A &= c.Mem.Get(addr)
		c.setNVflags(c.A)

	case 0x35: // AND zp,X
		c.A &= c.Mem.Get(uint16(byte(c.popByte() + c.X)))
		c.setNVflags(c.A)

	case 0x36: // ROL zp,X
		c.rolMem(uint16(byte(c.popByte() + c.X)))

	case 0x38: // SEC
		c.sec()

	case 0x39: // AND abs,Y
		c.A &= c.Mem.Get(c.popWord() + uint16(c.Y))
		c.setNVflags(c.A)

	case 0x3d: // AND abs,X
		c.A &= c.Mem.Get(c.popWord() + uint16(c.X))
		c.setNVflags(c.A)

	case 0x3e: // ROL abs,X
		c.rolMem(c.popWord() + uint16(c.X))

	case 0x40: // RTI
		c.P = c.stackPop() | 0x30
		lo := uint16(c.stackPop())
		hi := uint16(c.stackPop())
		c.PC = lo | hi<<8

	case 0x41: // EOR (zp,X)
		zp := byte(c.popByte() + c.X)
		addr := c.Mem.GetWord(uint16(zp))
		c.A ^= c.Mem.Get(addr)
		c.setNVflags(c.A)

	case 0x42: // WDM (pseudo): print A to messages when arg is 0
		v := c.popByte()
		if v == 0 {
			c.message(string([]byte{c.A}))
		}

	case 0x45: // EOR zp
		c.A ^= c.Mem.Get(uint16(c.popByte()))
		c.setNVflags(c.A)

	case 0x46: // LSR zp
		c.shiftRightMem(uint16(c.popByte()))

	case 0x48: // PHA
		c.stackPush(c.A)

	case 0x49: // EOR #imm
		c.A ^= c.popByte()
		c.setNVflags(c.A)

	case 0x4a: // LSR A
		c.setCarryFromBit0(c.A)
		c.A = c.A >> 1
		c.setNVflags(c.A)

	case 0x4c: // JMP abs
		c.PC = c.popWord()

	case 0x4d: // EOR abs
		c.A ^= c.Mem.Get(c.popWord())
		c.setNVflags(c.A)

	case 0x4e: // LSR abs
		c.shiftRightMem(c.popWord())

	case 0x50: // BVC
		off := c.popByte()
		if !c.overflowSet() {
			c.jumpBranch(off)
		}

	case 0x51: // EOR (zp),Y
		zp := c.popByte()
		addr := c.Mem.GetWord(uint16(zp)) + uint16(c.Y)
		c.A ^= c.Mem.Get(addr)
		c.setNVflags(c.A)

	case 0x55: // EOR zp,X
		c.A ^= c.Mem.Get(uint16(byte(c.popByte() + c.X)))
		c.setNVflags(c.A)

	case 0x56: // LSR zp,X
		c.shiftRightMem(uint16(byte(c.popByte() + c.X)))

	case 0x58: // CLI
		c.P &^= FlagI
		c.message("Interrupts not implemented")
		c.Running = false

	case 0x59: // EOR abs,Y
		c.A ^= c.Mem.Get(c.popWord() + uint16(c.Y))
		c.setNVflags(c.A)

	case 0x5d: // EOR abs,X
		c.A ^= c.Mem.Get(c.popWord() + uint16(c.X))
		c.setNVflags(c.A)

	case 0x5e: // LSR abs,X
		c.shiftRightMem(c.popWord() + uint16(c.X))

	case 0x60: // RTS
		lo := uint16(c.stackPop())
		hi := uint16(c.stackPop())
		c.PC = (lo | hi<<8) + 1

	case 0x61: // ADC (zp,X)
		zp := byte(c.popByte() + c.X)
		addr := c.Mem.GetWord(uint16(zp))
		c.testADC(c.Mem.Get(addr))

	case 0x65: // ADC zp
		c.testADC(c.Mem.Get(uint16(c.popByte())))

	case 0x66: // ROR zp
		c.rorMem(uint16(c.popByte()))

	case 0x68: // PLA
		c.A = c.stackPop()
		c.setNVflags(c.A)

	case 0x69: // ADC #imm
		c.testADC(c.popByte())

	case 0x6a: // ROR A
		sf := c.P & FlagC
		c.setCarryFromBit0(c.A)
		c.A = c.A >> 1
		if sf != 0 {
			c.A |= 0x80
		}
		c.setNVflags(c.A)

	case 0x6c: // JMP (abs)
		c.PC = c.Mem.GetWord(c.popWord())

	case 0x6d: // ADC abs
		c.testADC(c.Mem.Get(c.popWord()))

	case 0x6e: // ROR abs
		c.rorMem(c.popWord())

	case 0x70: // BVS
		off := c.popByte()
		if c.overflowSet() {
			c.jumpBranch(off)
		}

	case 0x71: // ADC (zp),Y
		zp := c.popByte()
		addr := c.Mem.GetWord(uint16(zp))
		c.testADC(c.Mem.Get(addr + uint16(c.Y)))

	case 0x75: // ADC zp,X
		c.testADC(c.Mem.Get(uint16(byte(c.popByte() + c.X))))

	case 0x76: // ROR zp,X
		c.rorMem(uint16(byte(c.popByte() + c.X)))

	case 0x78: // SEI
		c.P |= FlagI
		c.message("Interrupts not implemented")
		c.Running = false

	case 0x79: // ADC abs,Y
		addr := c.popWord()
		c.testADC(c.Mem.Get(addr + uint16(c.Y)))

	case 0x7d: // ADC abs,X
		addr := c.popWord()
		c.testADC(c.Mem.Get(addr + uint16(c.X)))

	case 0x7e: // ROR abs,X
		c.rorMem(c.popWord() + uint16(c.X))

	case 0x81: // STA (zp,X)
		zp := byte(c.popByte() + c.X)
		addr := c.Mem.GetWord(uint16(zp))
		c.Mem.StoreByte(addr, c.A)

	case 0x84: // STY zp
		c.Mem.StoreByte(uint16(c.popByte()), c.Y)

	case 0x85: // STA zp
		c.Mem.StoreByte(uint16(c.popByte()), c.A)

	case 0x86: // STX zp
		c.Mem.StoreByte(uint16(c.popByte()), c.X)

	case 0x88: // DEY
		c.Y--
		c.setNVflags(c.Y)

	case 0x8a: // TXA
		c.A = c.X
		c.setNVflags(c.A)

	case 0x8c: // STY abs
		c.Mem.StoreByte(c.popWord(), c.Y)

	case 0x8d: // STA abs
		c.Mem.StoreByte(c.popWord(), c.A)

	case 0x8e: // STX abs
		c.Mem.StoreByte(c.popWord(), c.X)

	case 0x90: // BCC
		off := c.popByte()
		if !c.carrySet() {
			c.jumpBranch(off)
		}

	case 0x91: // STA (zp),Y
		zp := c.popByte()
		addr := c.Mem.GetWord(uint16(zp)) + uint16(c.Y)
		c.Mem.StoreByte(addr, c.A)

	case 0x94: // STY zp,X
		c.Mem.StoreByte(uint16(byte(c.popByte()+c.X)), c.Y)

	case 0x95: // STA zp,X
		c.Mem.StoreByte(uint16(byte(c.popByte()+c.X)), c.A)

	case 0x96: // STX zp,Y
		c.Mem.StoreByte(uint16(byte(c.popByte()+c.Y)), c.X)

	case 0x98: // TYA
		c.A = c.Y
		c.setNVflags(c.A)

	case 0x99: // STA abs,Y
		c.Mem.StoreByte(c.popWord()+uint16(c.Y), c.A)

	case 0x9a: // TXS
		c.SP = c.X

	case 0x9d: // STA abs,X
		c.Mem.StoreByte(c.popWord()+uint16(c.X), c.A)

	case 0xa0: // LDY #imm
		c.Y = c.popByte()
		c.setNVflags(c.Y)

	case 0xa1: // LDA (zp,X)
		zp := byte(c.popByte() + c.X)
		addr := c.Mem.GetWord(uint16(zp))
		c.A = c.Mem.Get(addr)
		c.setNVflags(c.A)

	case 0xa2: // LDX #imm
		c.X = c.popByte()
		c.setNVflags(c.X)

	case 0xa4: // LDY zp
		c.Y = c.Mem.Get(uint16(c.popByte()))
		c.setNVflags(c.Y)

	case 0xa5: // LDA zp
		c.A = c.Mem.Get(uint16(c.popByte()))
		c.setNVflags(c.A)

	case 0xa6: // LDX zp
		c.X = c.Mem.Get(uint16(c.popByte()))
		c.setNVflags(c.X)

	case 0xa8: // TAY
		c.Y = c.A
		c.setNVflags(c.Y)

	case 0xa9: // LDA #imm
		c.A = c.popByte()
		c.setNVflags(c.A)

	case 0xaa: // TAX
		c.X = c.A
		c.setNVflags(c.X)

	case 0xac: // LDY abs
		c.Y = c.Mem.Get(c.popWord())
		c.setNVflags(c.Y)

	case 0xad: // LDA abs
		c.A = c.Mem.Get(c.popWord())
		c.setNVflags(c.A)

	case 0xae: // LDX abs
		c.X = c.Mem.Get(c.popWord())
		c.setNVflags(c.X)

	case 0xb0: // BCS
		off := c.popByte()
		if c.carrySet() {
			c.jumpBranch(off)
		}

	case 0xb1: // LDA (zp),Y
		zp := c.popByte()
		addr := c.Mem.GetWord(uint16(zp)) + uint16(c.Y)
		c.A = c.Mem.Get(addr)
		c.setNVflags(c.A)

	case 0xb4: // LDY zp,X
		c.Y = c.Mem.Get(uint16(byte(c.popByte() + c.X)))
		c.setNVflags(c.Y)

	case 0xb5: // LDA zp,X
		c.A = c.Mem.Get(uint16(byte(c.popByte() + c.X)))
		c.setNVflags(c.A)

	case 0xb6: // LDX zp,Y
		c.X = c.Mem.Get(uint16(byte(c.popByte() + c.Y)))
		c.setNVflags(c.X)

	case 0xb8: // CLV
		c.clv()

	case 0xb9: // LDA abs,Y
		c.A = c.Mem.Get(c.popWord() + uint16(c.Y))
		c.setNVflags(c.A)

	case 0xba: // TSX
		c.X = c.SP
		c.setNVflags(c.X)

	case 0xbc: // LDY abs,X
		c.Y = c.Mem.Get(c.popWord() + uint16(c.X))
		c.setNVflags(c.Y)

	case 0xbd: // LDA abs,X
		c.A = c.Mem.Get(c.popWord() + uint16(c.X))
		c.setNVflags(c.A)

	case 0xbe: // LDX abs,Y
		c.X = c.Mem.Get(c.popWord() + uint16(c.Y))
		c.setNVflags(c.X)

	case 0xc0: // CPY #imm
		c.doCompare(c.Y, c.popByte())

	case 0xc1: // CMP (zp,X)
		zp := byte(c.popByte() + c.X)
		addr := c.Mem.GetWord(uint16(zp))
		c.doCompare(c.A, c.Mem.Get(addr))

	case 0xc4: // CPY zp
		c.doCompare(c.Y, c.Mem.Get(uint16(c.popByte())))

	case 0xc5: // CMP zp
		c.doCompare(c.A, c.Mem.Get(uint16(c.popByte())))

	case 0xc6: // DEC zp
		c.decMem(uint16(c.popByte()))

	case 0xc8: // INY
		c.Y++
		c.setNVflags(c.Y)

	case 0xc9: // CMP #imm
		c.doCompare(c.A, c.popByte())

	case 0xca: // DEX
		c.X--
		c.setNVflags(c.X)

	case 0xcc: // CPY abs
		c.doCompare(c.Y, c.Mem.Get(c.popWord()))

	case 0xcd: // CMP abs
		c.doCompare(c.A, c.Mem.Get(c.popWord()))

	case 0xce: // DEC abs
		c.decMem(c.popWord())

	case 0xd0: // BNE
		off := c.popByte()
		if !c.zeroSet() {
			c.jumpBranch(off)
		}

	case 0xd1: // CMP (zp),Y
		zp := c.popByte()
		addr := c.Mem.GetWord(uint16(zp)) + uint16(c.Y)
		c.doCompare(c.A, c.Mem.Get(addr))

	case 0xd5: // CMP zp,X
		c.doCompare(c.A, c.Mem.Get(uint16(byte(c.popByte()+c.X))))

	case 0xd6: // DEC zp,X
		c.decMem(uint16(byte(c.popByte() + c.X)))

	case 0xd8: // CLD
		c.P &^= FlagD

	case 0xd9: // CMP abs,Y
		c.doCompare(c.A, c.Mem.Get(c.popWord()+uint16(c.Y)))

	case 0xdd: // CMP abs,X
		c.doCompare(c.A, c.Mem.Get(c.popWord()+uint16(c.X)))

	case 0xde: // DEC abs,X
		c.decMem(c.popWord() + uint16(c.X))

	case 0xe0: // CPX #imm
		c.doCompare(c.X, c.popByte())

	case 0xe1: // SBC (zp,X)
		zp := byte(c.popByte() + c.X)
		addr := c.Mem.GetWord(uint16(zp))
		c.testSBC(c.Mem.Get(addr))

	case 0xe4: // CPX zp
		c.doCompare(c.X, c.Mem.Get(uint16(c.popByte())))

	case 0xe5: // SBC zp
		c.testSBC(c.Mem.Get(uint16(c.popByte())))

	case 0xe6: // INC zp
		c.incMem(uint16(c.popByte()))

	case 0xe8: // INX
		c.X++
		c.setNVflags(c.X)

	case 0xe9: // SBC #imm
		c.testSBC(c.popByte())

	case 0xea: // NOP

	case 0xec: // CPX abs
		c.doCompare(c.X, c.Mem.Get(c.popWord()))

	case 0xed: // SBC abs
		c.testSBC(c.Mem.Get(c.popWord()))

	case 0xee: // INC abs
		c.incMem(c.popWord())

	case 0xf0: // BEQ
		off := c.popByte()
		if c.zeroSet() {
			c.jumpBranch(off)
		}

	case 0xf1: // SBC (zp),Y
		zp := c.popByte()
		addr := c.Mem.GetWord(uint16(zp))
		c.testSBC(c.Mem.Get(addr + uint16(c.Y)))

	case 0xf5: // SBC zp,X
		c.testSBC(c.Mem.Get(uint16(byte(c.popByte() + c.X))))

	case 0xf6: // INC zp,X
		c.incMem(uint16(byte(c.popByte() + c.X)))

	case 0xf8: // SED
		c.P |= FlagD

	case 0xf9: // SBC abs,Y
		addr := c.popWord()
		c.testSBC(c.Mem.Get(addr + uint16(c.Y)))

	case 0xfd: // SBC abs,X
		addr := c.popWord()
		c.testSBC(c.Mem.Get(addr + uint16(c.X)))

	case 0xfe: // INC abs,X
		c.incMem(c.popWord() + uint16(c.X))

	default:
		c.message("Address $" + hex16(c.PC-1) + " - unknown opcode")
		c.Running = false
	}
}

func hex16(v uint16) string {
	const d = "0123456789abcdef"
	return string([]byte{
		d[(v>>12)&0xf], d[(v>>8)&0xf], d[(v>>4)&0xf], d[v&0xf],
	})
}
