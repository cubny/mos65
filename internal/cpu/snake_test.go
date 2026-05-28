package cpu

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cubny/mos65/internal/assembler"
	"github.com/cubny/mos65/internal/memory"
)

// TestSnakeRunsBriefly assembles snake.asm and steps the CPU for a few thousand
// instructions to ensure no panics or unknown opcodes.
func TestSnakeRunsBriefly(t *testing.T) {
	path, _ := filepath.Abs("../../testdata/snake.asm")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("snake.asm not found: %v", err)
	}
	mem := memory.New()
	if _, err := assembler.Assemble(mem, string(data)); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	pixelWrites := 0
	mem.OnPixel = func(uint16) { pixelWrites++ }
	c := New(mem)
	seq := byte(0x11) // deterministic-ish; high enough that apple isn't pinned to a wall
	c.Rand = func() byte { seq = seq*7 + 1; return seq }
	c.Running = true
	for i := 0; i < 5000 && c.Running && c.PC != 0; i++ {
		c.Step()
	}
	t.Logf("after run: A=%02x X=%02x Y=%02x PC=$%04x pixel writes=%d",
		c.A, c.X, c.Y, c.PC, pixelWrites)
}
