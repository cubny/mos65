package memory

import (
	"fmt"
	"strings"
)

const (
	Size           = 0x10000
	ScreenStart    = 0x0200
	ScreenEnd      = 0x05ff
	RandomAddr     = 0x00fe
	KeyboardAddr   = 0x00ff
	ProgramStart   = 0x0600
)

type Memory struct {
	data    [Size]byte
	OnPixel func(addr uint16)
}

func New() *Memory {
	return &Memory{}
}

func (m *Memory) Get(addr uint16) byte {
	return m.data[addr]
}

func (m *Memory) Set(addr uint16, value byte) {
	m.data[addr] = value
}

func (m *Memory) GetWord(addr uint16) uint16 {
	return uint16(m.data[addr]) | uint16(m.data[addr+1])<<8
}

func (m *Memory) StoreByte(addr uint16, value byte) {
	m.data[addr] = value
	if addr >= ScreenStart && addr <= ScreenEnd && m.OnPixel != nil {
		m.OnPixel(addr)
	}
}

func (m *Memory) Clear(start, end int) {
	for i := start; i < end; i++ {
		m.data[i] = 0
	}
}

func (m *Memory) Format(start, length int) string {
	var b strings.Builder
	for x := 0; x < length; x++ {
		if x&15 == 0 {
			if x > 0 {
				b.WriteByte('\n')
			}
			fmt.Fprintf(&b, "%04x: ", start+x)
		}
		fmt.Fprintf(&b, "%02x ", m.data[start+x])
	}
	return b.String()
}
