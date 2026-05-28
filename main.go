package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cubny/mos65/internal/tui"
)

func main() {
	flag.Parse()

	var src string
	var filename string
	if flag.NArg() > 0 {
		filename = flag.Arg(0)
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read:", err)
			os.Exit(1)
		}
		src = string(data)
	}

	m := tui.New(src, filename)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tea:", err)
		os.Exit(1)
	}
}
