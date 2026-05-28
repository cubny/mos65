package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette is the 16-color palette from the JS Display module — pixels only.
var Palette = [16]lipgloss.Color{
	"#000000", "#ffffff", "#880000", "#aaffee",
	"#cc44cc", "#00cc55", "#0000aa", "#eeee77",
	"#dd8855", "#664400", "#ff7777", "#333333",
	"#777777", "#aaff66", "#0088ff", "#bbbbbb",
}

// UI design tokens — Charm-style.
const (
	colorAccent    = lipgloss.Color("#5FF7FF")
	colorAccentAlt = lipgloss.Color("#FFB454")
	colorMuted     = lipgloss.Color("#5C5C5C")
	colorText      = lipgloss.Color("#E6E6E6")
	colorDim       = lipgloss.Color("#888888")
	colorErr       = lipgloss.Color("#FF5F5F")
	colorOk        = lipgloss.Color("#A7E22E")
	colorWarn      = lipgloss.Color("#FFB454")
	colorBlack     = lipgloss.Color("#0B0B14")
)

const pillWidth = 5

var (
	titleChip = lipgloss.NewStyle().
			Background(colorAccent).
			Foreground(colorBlack).
			Bold(true).
			Padding(0, 1)

	hintStyle  = lipgloss.NewStyle().Foreground(colorDim)
	errorStyle = lipgloss.NewStyle().Foreground(colorErr).Bold(true)
	okStyle    = lipgloss.NewStyle().Foreground(colorOk)
	dimStyle   = lipgloss.NewStyle().Foreground(colorMuted).Italic(true)

	edgeFocused = lipgloss.NewStyle().Foreground(colorAccent)
	edgeBlurred = lipgloss.NewStyle().Foreground(colorMuted)

	titleFocused = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	titleBlurred = lipgloss.NewStyle().Foreground(colorMuted)

	pillOn = lipgloss.NewStyle().
		Width(pillWidth).
		Align(lipgloss.Center).
		Foreground(colorBlack).
		Bold(true)

	pillOff = lipgloss.NewStyle().
		Width(pillWidth).
		Align(lipgloss.Center).
		Foreground(colorMuted)
)

// chip renders a status pill of fixed width so toggling on/off doesn't shift
// horizontal alignment.
func chip(label string, fg lipgloss.Color, active bool) string {
	if !active {
		return pillOff.Render(label)
	}
	return pillOn.Background(fg).Render(label)
}

// renderBox draws a rounded box width-wide with title embedded in the top
// border (Charm style). Body may contain ANSI; padding uses lipgloss.Width so
// escape sequences don't break alignment.
func renderBox(title, body string, width int, focused bool) string {
	if width < 6 {
		width = 6
	}
	edge := edgeBlurred
	tStyle := titleBlurred
	if focused {
		edge = edgeFocused
		tStyle = titleFocused
	}

	titleText := " " + title + " "
	titleW := lipgloss.Width(titleText)
	const leftDashes = 2
	rightDashes := width - 2 - leftDashes - titleW
	if rightDashes < 0 {
		rightDashes = 0
	}
	top := edge.Render("╭"+strings.Repeat("─", leftDashes)) +
		tStyle.Render(titleText) +
		edge.Render(strings.Repeat("─", rightDashes)+"╮")
	bot := edge.Render("╰" + strings.Repeat("─", width-2) + "╯")

	inner := width - 4 // 2 border + 2 inner padding
	if inner < 1 {
		inner = 1
	}
	var rows []string
	rows = append(rows, top)
	for _, line := range strings.Split(body, "\n") {
		pad := inner - lipgloss.Width(line)
		if pad < 0 {
			line = truncate(line, inner)
			pad = 0
		}
		rows = append(rows,
			edge.Render("│")+" "+line+strings.Repeat(" ", pad)+" "+edge.Render("│"))
	}
	rows = append(rows, bot)
	return strings.Join(rows, "\n")
}

// truncate cuts a (possibly ANSI-styled) string to at most n display columns.
// Used as a safety net for lines that overflow a pane.
func truncate(s string, n int) string {
	if lipgloss.Width(s) <= n {
		return s
	}
	// Rough byte-based truncation; safe because all our overflowing strings
	// are plain ASCII (line numbers, hex dump). Styled multi-byte content
	// already fits its pane by construction.
	if n <= 0 {
		return ""
	}
	if n > len(s) {
		return s
	}
	return s[:n]
}
