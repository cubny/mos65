package assembler

import (
	"testing"

	"github.com/cubny/mos65/internal/memory"
)

func mustAssemble(t *testing.T, src string) (*memory.Memory, Result) {
	t.Helper()
	mem := memory.New()
	r, err := Assemble(mem, src)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	return mem, r
}

func bytesAt(mem *memory.Memory, start uint16, n int) []byte {
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = mem.Get(start + uint16(i))
	}
	return out
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestImmediateAndStore(t *testing.T) {
	mem, r := mustAssemble(t, "LDA #$01\nSTA $0200\n")
	want := []byte{0xa9, 0x01, 0x8d, 0x00, 0x02}
	if got := bytesAt(mem, 0x600, len(want)); !equal(got, want) {
		t.Fatalf("got % x, want % x", got, want)
	}
	if r.CodeLen != 5 {
		t.Fatalf("CodeLen = %d, want 5", r.CodeLen)
	}
}

func TestZeroPage(t *testing.T) {
	mem, _ := mustAssemble(t, "LDA $42\nSTA $43\n")
	want := []byte{0xa5, 0x42, 0x85, 0x43}
	if got := bytesAt(mem, 0x600, len(want)); !equal(got, want) {
		t.Fatalf("got % x, want % x", got, want)
	}
}

func TestIndexedAndIndirect(t *testing.T) {
	src := "LDA $42,X\nLDA $1234,Y\nLDA ($10,X)\nLDA ($10),Y\nJMP ($abcd)\n"
	mem, _ := mustAssemble(t, src)
	want := []byte{
		0xb5, 0x42,
		0xb9, 0x34, 0x12,
		0xa1, 0x10,
		0xb1, 0x10,
		0x6c, 0xcd, 0xab,
	}
	if got := bytesAt(mem, 0x600, len(want)); !equal(got, want) {
		t.Fatalf("got % x, want % x", got, want)
	}
}

func TestLabelsAndBranch(t *testing.T) {
	src := "loop:\n  LDA #$01\n  BNE loop\n"
	mem, _ := mustAssemble(t, src)
	// 0x600: A9 01 D0 FC (offset -4 from PC after BNE = 0x604, target 0x600 -> distance -4)
	want := []byte{0xa9, 0x01, 0xd0, 0xfc}
	if got := bytesAt(mem, 0x600, len(want)); !equal(got, want) {
		t.Fatalf("got % x, want % x", got, want)
	}
}

func TestForwardLabelAbsolute(t *testing.T) {
	src := "  JMP later\n  NOP\nlater:\n  NOP\n"
	mem, _ := mustAssemble(t, src)
	// JMP $0604 (origin+3+1 = 0x604), NOP, NOP
	want := []byte{0x4c, 0x04, 0x06, 0xea, 0xea}
	if got := bytesAt(mem, 0x600, len(want)); !equal(got, want) {
		t.Fatalf("got % x, want % x", got, want)
	}
}

func TestDCBAndDefine(t *testing.T) {
	src := "define COUNT 5\n  LDX #COUNT\n  DCB $de,$ad,1,%00001111\n"
	mem, _ := mustAssemble(t, src)
	want := []byte{0xa2, 0x05, 0xde, 0xad, 0x01, 0x0f}
	if got := bytesAt(mem, 0x600, len(want)); !equal(got, want) {
		t.Fatalf("got % x, want % x", got, want)
	}
}

func TestBranchOutOfRange(t *testing.T) {
	// 130 NOPs then BNE backward to start.
	src := "start:\n"
	for i := 0; i < 130; i++ {
		src += "  NOP\n"
	}
	src += "  BNE start\n"
	mem := memory.New()
	_, err := Assemble(mem, src)
	if err == nil {
		t.Fatal("expected out-of-range branch error")
	}
	ae, ok := err.(*Error)
	if !ok || !ae.OutOfRangeBranch {
		t.Fatalf("expected OutOfRangeBranch error, got %v", err)
	}
}

func TestImmediateLabelHiLo(t *testing.T) {
	src := "  LDA #<msg\n  LDX #>msg\n  RTS\nmsg:\n  DCB $48\n"
	mem, _ := mustAssemble(t, src)
	// LDA #<msg: msg lives at 0x605 -> lo=0x05; LDX #>msg: hi=0x06
	want := []byte{0xa9, 0x05, 0xa2, 0x06, 0x60, 0x48}
	if got := bytesAt(mem, 0x600, len(want)); !equal(got, want) {
		t.Fatalf("got % x, want % x", got, want)
	}
}

func TestOriginDirective(t *testing.T) {
	src := "  LDA #$01\n*=$0800\n  LDA #$02\n"
	mem, _ := mustAssemble(t, src)
	if got := mem.Get(0x600); got != 0xa9 {
		t.Fatalf("0x600 = %02x, want a9", got)
	}
	if got := mem.Get(0x601); got != 0x01 {
		t.Fatalf("0x601 = %02x, want 01", got)
	}
	if got := mem.Get(0x800); got != 0xa9 {
		t.Fatalf("0x800 = %02x, want a9", got)
	}
	if got := mem.Get(0x801); got != 0x02 {
		t.Fatalf("0x801 = %02x, want 02", got)
	}
}
