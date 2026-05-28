package cpu

import (
	"testing"

	"github.com/cubny/mos65/internal/assembler"
	"github.com/cubny/mos65/internal/memory"
)

func run(t *testing.T, src string) *CPU {
	t.Helper()
	mem := memory.New()
	if _, err := assembler.Assemble(mem, src); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	c := New(mem)
	c.Rand = func() byte { return 0 } // deterministic
	c.Running = true
	// Run until BRK halts or PC wraps to 0.
	for i := 0; i < 10000 && c.Running; i++ {
		c.Step()
		if c.PC == 0 {
			break
		}
	}
	return c
}

func TestLDA_STA(t *testing.T) {
	c := run(t, "LDA #$42\nSTA $0200\nBRK\n")
	if c.A != 0x42 {
		t.Fatalf("A=%02x want 42", c.A)
	}
	if c.Mem.Get(0x0200) != 0x42 {
		t.Fatalf("$0200=%02x want 42", c.Mem.Get(0x0200))
	}
}

func TestINX(t *testing.T) {
	c := run(t, "LDX #$00\nINX\nINX\nINX\nBRK\n")
	if c.X != 3 {
		t.Fatalf("X=%d want 3", c.X)
	}
}

func TestLoopAndBranch(t *testing.T) {
	src := `
  LDX #$0a
loop:
  DEX
  BNE loop
  BRK
`
	c := run(t, src)
	if c.X != 0 {
		t.Fatalf("X=%d want 0", c.X)
	}
	if c.P&FlagZ == 0 {
		t.Fatalf("Z flag not set, P=%08b", c.P)
	}
}

func TestADCBinary(t *testing.T) {
	c := run(t, "CLC\nLDA #$10\nADC #$20\nBRK\n")
	if c.A != 0x30 {
		t.Fatalf("A=%02x want 30", c.A)
	}
	if c.P&FlagC != 0 {
		t.Fatal("carry should be clear")
	}
}

func TestADCBinaryOverflow(t *testing.T) {
	c := run(t, "CLC\nLDA #$7f\nADC #$01\nBRK\n")
	if c.A != 0x80 {
		t.Fatalf("A=%02x want 80", c.A)
	}
	if c.P&FlagV == 0 {
		t.Fatal("overflow should be set (7f+1=80)")
	}
}

func TestADCDecimal(t *testing.T) {
	c := run(t, "SED\nCLC\nLDA #$19\nADC #$01\nCLD\nBRK\n")
	if c.A != 0x20 {
		t.Fatalf("A=%02x want 20 (BCD)", c.A)
	}
	if c.P&FlagC != 0 {
		t.Fatal("carry should be clear")
	}
}

func TestADCDecimalCarry(t *testing.T) {
	c := run(t, "SED\nCLC\nLDA #$99\nADC #$01\nCLD\nBRK\n")
	if c.A != 0x00 {
		t.Fatalf("A=%02x want 00 (BCD wrap)", c.A)
	}
	if c.P&FlagC == 0 {
		t.Fatal("carry should be set")
	}
}

func TestSBCBinary(t *testing.T) {
	c := run(t, "SEC\nLDA #$50\nSBC #$20\nBRK\n")
	if c.A != 0x30 {
		t.Fatalf("A=%02x want 30", c.A)
	}
	if c.P&FlagC == 0 {
		t.Fatal("carry should be set (no borrow)")
	}
}

func TestSBCDecimal(t *testing.T) {
	c := run(t, "SED\nSEC\nLDA #$50\nSBC #$25\nCLD\nBRK\n")
	if c.A != 0x25 {
		t.Fatalf("A=%02x want 25 (BCD)", c.A)
	}
}

func TestStackPushPop(t *testing.T) {
	c := run(t, "LDA #$aa\nPHA\nLDA #$00\nPLA\nBRK\n")
	if c.A != 0xaa {
		t.Fatalf("A=%02x want aa", c.A)
	}
}

func TestJSRRTS(t *testing.T) {
	src := `
  JSR sub
  LDX #$05
  BRK
sub:
  LDA #$11
  RTS
`
	c := run(t, src)
	if c.A != 0x11 {
		t.Fatalf("A=%02x want 11", c.A)
	}
	if c.X != 0x05 {
		t.Fatalf("X=%02x want 05", c.X)
	}
}

func TestScreenWritesTriggerPixel(t *testing.T) {
	mem := memory.New()
	if _, err := assembler.Assemble(mem, "LDA #$05\nSTA $0200\nBRK\n"); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	var pixels []uint16
	mem.OnPixel = func(addr uint16) { pixels = append(pixels, addr) }
	c := New(mem)
	c.Rand = func() byte { return 0 }
	c.Running = true
	for i := 0; i < 100 && c.Running; i++ {
		c.Step()
	}
	if len(pixels) != 1 || pixels[0] != 0x0200 {
		t.Fatalf("pixels=%v want [0x0200]", pixels)
	}
}
