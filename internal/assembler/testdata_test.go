package assembler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cubny/mos65/internal/memory"
)

func TestAssembleTestdataPrograms(t *testing.T) {
	root, _ := filepath.Abs("../../testdata")
	files, err := os.ReadDir(root)
	if err != nil {
		t.Skipf("no testdata directory: %v", err)
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".asm" {
			continue
		}
		t.Run(f.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, f.Name()))
			if err != nil {
				t.Fatal(err)
			}
			mem := memory.New()
			r, err := Assemble(mem, string(data))
			if err != nil {
				t.Fatalf("%s failed to assemble: %v", f.Name(), err)
			}
			if r.CodeLen == 0 {
				t.Fatalf("%s assembled to 0 bytes", f.Name())
			}
		})
	}
}
