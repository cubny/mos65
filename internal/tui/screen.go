package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderScreen draws a 32x32 framebuffer as 16 rows of 32 half-block cells.
// Each cell shows two stacked pixels: the top pixel as foreground, the bottom
// as background, of the U+2580 "upper half block" glyph.
func renderScreen(fb [32][32]byte) string {
	var b strings.Builder
	for y := 0; y < 32; y += 2 {
		for x := 0; x < 32; x++ {
			top := fb[y][x] & 0x0f
			bot := fb[y+1][x] & 0x0f
			style := lipgloss.NewStyle().
				Foreground(Palette[top]).
				Background(Palette[bot])
			b.WriteString(style.Render("▀"))
		}
		if y+2 < 32 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
