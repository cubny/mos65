package cpu

func (c *CPU) carryInt() int {
	if c.carrySet() {
		return 1
	}
	return 0
}

func (c *CPU) testADC(value byte) {
	if (c.A^value)&0x80 != 0 {
		c.clv()
	} else {
		c.setOverflow()
	}

	var tmp int
	if c.decimalMode() {
		tmp = int(c.A&0x0f) + int(value&0x0f) + c.carryInt()
		if tmp >= 10 {
			tmp = 0x10 | ((tmp + 6) & 0x0f)
		}
		tmp += int(c.A&0xf0) + int(value&0xf0)
		if tmp >= 160 {
			c.sec()
			if c.overflowSet() && tmp >= 0x180 {
				c.clv()
			}
			tmp += 0x60
		} else {
			c.clc()
			if c.overflowSet() && tmp < 0x80 {
				c.clv()
			}
		}
	} else {
		tmp = int(c.A) + int(value) + c.carryInt()
		if tmp >= 0x100 {
			c.sec()
			if c.overflowSet() && tmp >= 0x180 {
				c.clv()
			}
		} else {
			c.clc()
			if c.overflowSet() && tmp < 0x80 {
				c.clv()
			}
		}
	}
	c.A = byte(tmp & 0xff)
	c.setNVflags(c.A)
}

func (c *CPU) testSBC(value byte) {
	if (c.A^value)&0x80 != 0 {
		c.setOverflow()
	} else {
		c.clv()
	}

	var w int
	if c.decimalMode() {
		tmp := 0x0f + int(c.A&0x0f) - int(value&0x0f) + c.carryInt()
		if tmp < 0x10 {
			w = 0
			tmp -= 6
		} else {
			w = 0x10
			tmp -= 0x10
		}
		w += 0xf0 + int(c.A&0xf0) - int(value&0xf0)
		if w < 0x100 {
			c.clc()
			if c.overflowSet() && w < 0x80 {
				c.clv()
			}
			w -= 0x60
		} else {
			c.sec()
			if c.overflowSet() && w >= 0x180 {
				c.clv()
			}
		}
		w += tmp
	} else {
		w = 0xff + int(c.A) - int(value) + c.carryInt()
		if w < 0x100 {
			c.clc()
			if c.overflowSet() && w < 0x80 {
				c.clv()
			}
		} else {
			c.sec()
			if c.overflowSet() && w >= 0x180 {
				c.clv()
			}
		}
	}
	c.A = byte(w & 0xff)
	c.setNVflags(c.A)
}

// shiftLeftMem performs ASL on the byte at addr.
func (c *CPU) shiftLeftMem(addr uint16) {
	v := c.Mem.Get(addr)
	c.setCarryFromBit7(v)
	v = v << 1
	c.Mem.StoreByte(addr, v)
	c.setNVflags(v)
}

// shiftRightMem performs LSR on the byte at addr.
func (c *CPU) shiftRightMem(addr uint16) {
	v := c.Mem.Get(addr)
	c.setCarryFromBit0(v)
	v = v >> 1
	c.Mem.StoreByte(addr, v)
	c.setNVflags(v)
}

// rolMem performs ROL on the byte at addr.
func (c *CPU) rolMem(addr uint16) {
	sf := c.P & FlagC
	v := c.Mem.Get(addr)
	c.setCarryFromBit7(v)
	v = (v << 1) | sf
	c.Mem.StoreByte(addr, v)
	c.setNVflags(v)
}

// rorMem performs ROR on the byte at addr.
func (c *CPU) rorMem(addr uint16) {
	sf := c.P & FlagC
	v := c.Mem.Get(addr)
	c.setCarryFromBit0(v)
	v = v >> 1
	if sf != 0 {
		v |= 0x80
	}
	c.Mem.StoreByte(addr, v)
	c.setNVflags(v)
}

// bit performs BIT on value vs A.
func (c *CPU) bit(value byte) {
	if value&0x80 != 0 {
		c.P |= FlagN
	} else {
		c.P &^= FlagN
	}
	if value&0x40 != 0 {
		c.P |= FlagV
	} else {
		c.P &^= FlagV
	}
	if c.A&value == 0 {
		c.P |= FlagZ
	} else {
		c.P &^= FlagZ
	}
}

func (c *CPU) decMem(addr uint16) {
	v := c.Mem.Get(addr) - 1
	c.Mem.StoreByte(addr, v)
	c.setNVflags(v)
}

func (c *CPU) incMem(addr uint16) {
	v := c.Mem.Get(addr) + 1
	c.Mem.StoreByte(addr, v)
	c.setNVflags(v)
}
