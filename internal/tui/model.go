package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cubny/mos65/internal/assembler"
	"github.com/cubny/mos65/internal/cpu"
	"github.com/cubny/mos65/internal/memory"
)

type pane int

const (
	paneEditor pane = iota
	paneScreen
	paneDebugger
	paneMonitorStart
	paneMonitorLength
	paneCount
)

var paneNames = [...]string{
	paneEditor:        "editor",
	paneScreen:        "screen",
	paneDebugger:      "debugger",
	paneMonitorStart:  "monitor:start",
	paneMonitorLength: "monitor:length",
}

// keyMap groups the bindings shown in the help bar.
type keyMap struct {
	Tab        key.Binding
	ShiftTab   key.Binding
	Esc        key.Binding
	Assemble   key.Binding
	Run        key.Binding
	Reset      key.Binding
	Step       key.Binding
	Goto       key.Binding
	Debug      key.Binding
	Monitor    key.Binding
	Hexdump    key.Binding
	Disasm     key.Binding
	EditExt    key.Binding
	Save       key.Binding
	Notes      key.Binding
	HelpToggle key.Binding
	Quit       key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Tab:        key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "focus")),
		ShiftTab:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("⇧tab", "back")),
		Esc:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "→editor")),
		Assemble:   key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("^A", "assemble")),
		Run:        key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("^R", "run/stop")),
		Reset:      key.NewBinding(key.WithKeys("ctrl+t"), key.WithHelp("^T", "reset")),
		Step:       key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "step")),
		Goto:       key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "goto")),
		Debug:      key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("^D", "debug")),
		Monitor:    key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("^B", "monitor")),
		Hexdump:    key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("^F", "hexdump")),
		Disasm:     key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("^L", "disasm")),
		EditExt:    key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("^E", "$EDITOR")),
		Save:       key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("^S", "save")),
		Notes:      key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("^N", "notes")),
		HelpToggle: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit:       key.NewBinding(key.WithKeys("ctrl+c", "ctrl+q"), key.WithHelp("^Q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Assemble, k.Run, k.Debug, k.Monitor, k.EditExt, k.HelpToggle, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.ShiftTab, k.Esc},
		{k.Assemble, k.Run, k.Reset, k.Step, k.Goto},
		{k.Debug, k.Monitor, k.Hexdump, k.Disasm},
		{k.EditExt, k.Save, k.Notes, k.HelpToggle, k.Quit},
	}
}

// Model is the root Bubble Tea model.
type Model struct {
	width, height            int
	editorW, editorH, rightW int
	focus                    pane

	mem *memory.Memory
	cpu *cpu.CPU

	editor      textarea.Model
	messages    viewport.Model
	monitorView viewport.Model
	startInput  textinput.Model
	lenInput    textinput.Model
	gotoInput   textinput.Model
	gotoActive  bool

	help help.Model
	keys keyMap

	framebuffer [32][32]byte

	assembled bool
	codeLen   int
	running   bool
	debug     bool
	monitorOn bool

	msgBuf   strings.Builder
	filename string
}

// New constructs a Model with the given optional source preloaded into the editor.
func New(initialSrc string, filename string) *Model {
	mem := memory.New()
	c := cpu.New(mem)

	ed := textarea.New()
	ed.SetValue(initialSrc)
	ed.Placeholder = "; type 6502 assembly here — Ctrl+E for $EDITOR"
	ed.Focus()
	ed.CharLimit = 0
	ed.ShowLineNumbers = true
	ed.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2a"))
	ed.BlurredStyle.Base = lipgloss.NewStyle().Foreground(colorMuted)
	ed.BlurredStyle.Text = lipgloss.NewStyle().Foreground(colorMuted)
	ed.BlurredStyle.LineNumber = lipgloss.NewStyle().Foreground(colorMuted)
	ed.BlurredStyle.CursorLineNumber = lipgloss.NewStyle().Foreground(colorMuted)

	vp := viewport.New(40, 5)
	mv := viewport.New(40, 8)

	si := textinput.New()
	si.Prompt = "start $"
	si.SetValue("0")
	si.CharLimit = 4
	si.Width = 6

	li := textinput.New()
	li.Prompt = "length $"
	li.SetValue("ff")
	li.CharLimit = 4
	li.Width = 6

	gi := textinput.New()
	gi.Prompt = "goto: "
	gi.Width = 16

	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(colorDim)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(colorMuted)
	h.Styles.FullKey = h.Styles.ShortKey
	h.Styles.FullDesc = h.Styles.ShortDesc
	h.Styles.FullSeparator = h.Styles.ShortSeparator
	h.ShortSeparator = "  ·  "

	m := &Model{
		mem:         mem,
		cpu:         c,
		editor:      ed,
		messages:    vp,
		monitorView: mv,
		startInput:  si,
		lenInput:    li,
		gotoInput:   gi,
		focus:       paneEditor,
		filename:    filename,
		help:        h,
		keys:        newKeyMap(),
	}
	mem.OnPixel = m.onPixel
	c.OnMessage = m.appendMessage
	return m
}

// setFocus updates focus and synchronizes Focus()/Blur() side effects on the
// underlying bubble components.
func (m *Model) setFocus(p pane) {
	m.focus = p
	if p == paneEditor {
		m.editor.Focus()
	} else {
		m.editor.Blur()
	}
	if p == paneMonitorStart {
		m.startInput.Focus()
	} else {
		m.startInput.Blur()
	}
	if p == paneMonitorLength {
		m.lenInput.Focus()
	} else {
		m.lenInput.Blur()
	}
}

// onPixel must be a method value bound to a Model field, not the Model itself
// (Model is copied between Update calls). We solve this by routing through a
// closure set in Init.
func (m *Model) onPixel(addr uint16) {
	rel := int(addr) - memory.ScreenStart
	y := rel / 32
	x := rel % 32
	if y >= 0 && y < 32 && x >= 0 && x < 32 {
		m.framebuffer[y][x] = m.mem.Get(addr) & 0x0f
	}
}

func (m *Model) appendMessage(s string) {
	if len(s) > 1 {
		s += "\n"
	}
	m.msgBuf.WriteString(s)
	m.messages.SetContent(m.msgBuf.String())
	m.messages.GotoBottom()
}

// tickMsg drives the run loop.
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(15*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// editorDoneMsg signals that the external editor exited.
type editorDoneMsg struct {
	err     error
	content string
	path    string
}

func (m *Model) Init() tea.Cmd {
	return tickCmd()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
		return m, nil

	case tickMsg:
		m.runTick()
		return m, tickCmd()

	case editorDoneMsg:
		if msg.err != nil {
			m.appendMessage("editor: " + msg.err.Error())
		} else {
			m.editor.SetValue(msg.content)
			if msg.path != "" {
				m.filename = msg.path
				m.appendMessage("loaded " + msg.path)
			}
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) layout() {
	editorW := m.width * 11 / 20 // 55%
	if editorW > 90 {
		editorW = 90
	}
	if editorW < 50 {
		editorW = 50
	}
	if editorW > m.width-40 {
		editorW = m.width - 40
	}
	rightW := m.width - editorW
	if rightW < 40 {
		rightW = 40
	}
	editorH := m.height - 12 // header(1) + messages(~7) + help(2) + margin
	if editorH < 12 {
		editorH = 12
	}

	m.editorW = editorW
	m.editorH = editorH
	m.rightW = rightW

	// Pane content area = outer - 2 (border) - 2 (inner padding).
	m.editor.SetWidth(editorW - 4)
	m.editor.SetHeight(editorH - 2)
	m.help.Width = m.width
	m.messages.Width = m.width - 4
	m.messages.Height = 5
	m.monitorView.Width = rightW - 4
	m.monitorView.Height = 4
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Goto prompt overlay captures all keys until done.
	if m.gotoActive {
		switch msg.String() {
		case "esc":
			m.gotoActive = false
			m.gotoInput.SetValue("")
			return m, nil
		case "enter":
			m.applyGoto(m.gotoInput.Value())
			m.gotoActive = false
			m.gotoInput.SetValue("")
			return m, nil
		}
		var cmd tea.Cmd
		m.gotoInput, cmd = m.gotoInput.Update(msg)
		return m, cmd
	}

	// Focus cycling — must come before the editor dispatch so Tab is never
	// swallowed by bubbles/textarea.
	switch msg.String() {
	case "tab":
		m.setFocus((m.focus + 1) % paneCount)
		return m, nil
	case "shift+tab":
		m.setFocus((m.focus + paneCount - 1) % paneCount)
		return m, nil
	case "esc":
		if m.focus != paneEditor {
			m.setFocus(paneEditor)
			return m, nil
		}
	}

	// Global commands (work from any pane).
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Assemble):
		m.assemble()
		return m, nil
	case key.Matches(msg, m.keys.Run):
		m.toggleRun()
		return m, nil
	case key.Matches(msg, m.keys.Reset):
		m.reset()
		return m, nil
	case key.Matches(msg, m.keys.Hexdump):
		if m.assembled {
			m.appendMessage(assembler.Hexdump(m.mem, m.codeLen))
		}
		return m, nil
	case key.Matches(msg, m.keys.Disasm):
		if m.assembled {
			m.appendMessage(assembler.Disassemble(m.mem, m.codeLen))
		}
		return m, nil
	case key.Matches(msg, m.keys.Notes):
		m.showNotes()
		return m, nil
	case key.Matches(msg, m.keys.Debug):
		m.debug = !m.debug
		if m.debug {
			m.setFocus(paneDebugger)
			m.appendMessage("debugger ON — s steps, g goto, ^D to resume")
		} else {
			m.appendMessage("debugger OFF")
		}
		return m, nil
	case key.Matches(msg, m.keys.Monitor):
		m.monitorOn = !m.monitorOn
		if m.monitorOn {
			m.refreshMonitor()
		}
		return m, nil
	case key.Matches(msg, m.keys.Save):
		m.saveFile()
		return m, nil
	case key.Matches(msg, m.keys.EditExt):
		return m, m.editInExternal()
	case key.Matches(msg, m.keys.HelpToggle):
		m.help.ShowAll = !m.help.ShowAll
		return m, nil
	}

	// Pane-specific dispatch.
	switch m.focus {
	case paneEditor:
		var cmd tea.Cmd
		m.editor, cmd = m.editor.Update(msg)
		return m, cmd

	case paneScreen:
		if len(msg.Runes) > 0 {
			m.mem.Set(memory.KeyboardAddr, byte(msg.Runes[0]))
			if !m.running {
				m.appendMessage(fmt.Sprintf("key→$ff %02x", msg.Runes[0]))
			}
		}
		return m, nil

	case paneDebugger:
		switch {
		case key.Matches(msg, m.keys.Step):
			m.step()
		case key.Matches(msg, m.keys.Goto):
			m.gotoActive = true
			m.gotoInput.Focus()
			return m, textinput.Blink
		}
		return m, nil

	case paneMonitorStart:
		var cmd tea.Cmd
		m.startInput, cmd = m.startInput.Update(msg)
		m.refreshMonitor()
		return m, cmd

	case paneMonitorLength:
		var cmd tea.Cmd
		m.lenInput, cmd = m.lenInput.Update(msg)
		m.refreshMonitor()
		return m, cmd
	}
	return m, nil
}

func (m *Model) assemble() {
	src := m.editor.Value()
	m.cpu.Reset()
	r, err := assembler.Assemble(m.mem, src)
	if err != nil {
		m.appendMessage("**" + err.Error() + "**")
		m.assembled = false
		return
	}
	m.codeLen = r.CodeLen
	m.assembled = true
	m.running = false
	m.clearFramebuffer()
	m.appendMessage(fmt.Sprintf("Code assembled successfully, %d bytes.", r.CodeLen))
}

func (m *Model) toggleRun() {
	if !m.assembled {
		m.appendMessage("Assemble first (Ctrl+A).")
		return
	}
	if m.running {
		m.running = false
		m.cpu.Running = false
		m.appendMessage("\nStopped\n")
		return
	}
	m.running = true
	m.cpu.Running = true
}

func (m *Model) reset() {
	m.cpu.Reset()
	m.running = false
	m.clearFramebuffer()
	m.appendMessage("reset.")
}

func (m *Model) clearFramebuffer() {
	for y := range m.framebuffer {
		for x := range m.framebuffer[y] {
			m.framebuffer[y][x] = 0
		}
	}
}

func (m *Model) runTick() {
	if !m.running {
		if m.monitorOn {
			m.refreshMonitor()
		}
		return
	}
	if !m.debug {
		for i := 0; i < 97 && m.cpu.Running; i++ {
			m.cpu.Step()
			if m.cpu.PC == 0 {
				break
			}
		}
		if !m.cpu.Running || m.cpu.PC == 0 {
			m.running = false
			m.appendMessage(fmt.Sprintf("Program end at PC=$%04x", m.cpu.PC-1))
		}
	}
	if m.monitorOn {
		m.refreshMonitor()
	}
}

func (m *Model) step() {
	if !m.assembled {
		return
	}
	m.cpu.Running = true
	m.cpu.Step()
	if !m.cpu.Running || m.cpu.PC == 0 {
		m.running = false
		m.appendMessage(fmt.Sprintf("Program end at PC=$%04x", m.cpu.PC-1))
	}
}

func (m *Model) applyGoto(s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}
	if strings.HasPrefix(s, "$") {
		s = s[1:]
	} else if strings.HasPrefix(strings.ToLower(s), "0x") {
		s = s[2:]
	}
	var v uint64
	_, err := fmt.Sscanf(s, "%x", &v)
	if err != nil {
		m.appendMessage("Unable to parse address.")
		return
	}
	m.cpu.PC = uint16(v)
	m.appendMessage(fmt.Sprintf("PC=$%04x", m.cpu.PC))
}

func (m *Model) refreshMonitor() {
	startStr := strings.TrimSpace(m.startInput.Value())
	lenStr := strings.TrimSpace(m.lenInput.Value())
	var start, length uint64
	if _, err := fmt.Sscanf(startStr, "%x", &start); err != nil {
		m.monitorView.SetContent("invalid start")
		return
	}
	if _, err := fmt.Sscanf(lenStr, "%x", &length); err != nil {
		m.monitorView.SetContent("invalid length")
		return
	}
	end := start + length - 1
	if end > 0xffff || length == 0 {
		m.monitorView.SetContent("range out of bounds")
		return
	}
	m.monitorView.SetContent(m.mem.Format(int(start), int(length)))
}

func (m *Model) showNotes() {
	const notes = `Memory map:
  $fe          - new random byte each instruction
  $ff          - ASCII of last key pressed
  $200..$5ff   - 32x32 screen pixels (low nibble = color)

Palette (low nibble):
  0 black  1 white  2 red     3 cyan
  4 purple 5 green  6 blue    7 yellow
  8 orange 9 brown  a lt-red  b dk-gray
  c gray   d lt-grn e lt-blu  f lt-gray

Keys:
  Ctrl+A assemble    Ctrl+R run/stop   Ctrl+T reset
  Ctrl+D debug       Ctrl+B monitor    Ctrl+S save
  Ctrl+E edit in $EDITOR
  Ctrl+F hexdump     Ctrl+L disassemble
  Ctrl+N notes       Ctrl+Q quit
  Tab cycles every pane (including monitor sub-fields). Esc → editor.
  In Debugger pane: s=step, g=goto.
  In Screen pane: any keystroke writes its ASCII code to $ff (always).`
	m.appendMessage(notes)
}

func (m *Model) saveFile() {
	if m.filename == "" {
		m.appendMessage("No file loaded; use Ctrl+E to edit in $EDITOR.")
		return
	}
	if err := os.WriteFile(m.filename, []byte(m.editor.Value()), 0644); err != nil {
		m.appendMessage("save: " + err.Error())
		return
	}
	m.appendMessage("saved " + m.filename)
}

func (m *Model) editInExternal() tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	path := m.filename
	tempCreated := false
	if path == "" {
		f, err := os.CreateTemp("", "6502-*.asm")
		if err != nil {
			m.appendMessage("temp: " + err.Error())
			return nil
		}
		path = f.Name()
		f.Close()
		tempCreated = true
	}
	if err := os.WriteFile(path, []byte(m.editor.Value()), 0644); err != nil {
		m.appendMessage("write: " + err.Error())
		return nil
	}
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return editorDoneMsg{err: err}
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return editorDoneMsg{err: rerr}
		}
		msg := editorDoneMsg{content: string(data)}
		if !tempCreated {
			msg.path = path
		}
		return msg
	})
}

// View renders the whole UI.
func (m *Model) View() string {
	if m.width == 0 {
		return "initializing..."
	}

	header := m.renderHeader()

	editorPane := renderBox("editor", m.editor.View(), m.editorW, m.focus == paneEditor)

	screenInnerW := m.rightW - 4
	if screenInnerW < 32 {
		screenInnerW = 32
	}
	screenContent := lipgloss.Place(screenInnerW, 16, lipgloss.Center, lipgloss.Center,
		renderScreen(m.framebuffer))
	screenPane := renderBox("screen", screenContent, m.rightW, m.focus == paneScreen)

	debuggerPane := renderBox("debugger", m.renderDebuggerBody(), m.rightW, m.focus == paneDebugger)
	monitorFocused := m.focus == paneMonitorStart || m.focus == paneMonitorLength
	monitorPane := renderBox("monitor", m.renderMonitorBody(), m.rightW, monitorFocused)

	rightCol := lipgloss.JoinVertical(lipgloss.Left, screenPane, debuggerPane, monitorPane)
	body := lipgloss.JoinHorizontal(lipgloss.Top, editorPane, rightCol)

	messagesPane := renderBox("messages", m.messages.View(), m.width, false)

	footer := m.help.View(m.keys)

	out := lipgloss.JoinVertical(lipgloss.Left, header, body, messagesPane, footer)

	if m.gotoActive {
		overlay := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccentAlt).
			Padding(1, 2).
			Render(titleChip.Render("jump to") + "\n\n" +
				m.gotoInput.View() + "\n\n" +
				hintStyle.Render("enter ok  ·  esc cancel"))
		out = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay,
			lipgloss.WithWhitespaceChars(" "))
	}
	return out
}

func (m *Model) renderHeader() string {
	title := titleChip.Render("mos65")
	pills := chip("ASM", colorOk, m.assembled) + " " +
		chip("RUN", colorAccent, m.running) + " " +
		chip("DBG", colorAccentAlt, m.debug) + " " +
		chip("MON", colorAccentAlt, m.monitorOn)
	left := lipgloss.JoinHorizontal(lipgloss.Top, title, "  ", pills)

	var rightParts []string
	if m.assembled {
		rightParts = append(rightParts, fmt.Sprintf("%d bytes", m.codeLen))
	}
	rightParts = append(rightParts, paneNames[m.focus])
	if m.filename != "" {
		rightParts = append(rightParts, m.filename)
	}
	right := hintStyle.Render(strings.Join(rightParts, " · "))

	spacer := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if spacer < 1 {
		spacer = 1
	}
	return left + strings.Repeat(" ", spacer) + right
}

func (m *Model) renderDebuggerBody() string {
	if !m.debug {
		return m.cpu.DebugLine() + "\n" + dimStyle.Render("press ^D to enable stepping")
	}
	return m.cpu.DebugLine() + "\n" + hintStyle.Render("[s]tep  [g]oto")
}

func (m *Model) renderMonitorBody() string {
	if !m.monitorOn {
		return dimStyle.Render("monitor off — Ctrl+B to enable")
	}
	inputs := lipgloss.JoinHorizontal(lipgloss.Top,
		m.startInput.View(), "   ", m.lenInput.View())
	return inputs + "\n" + m.monitorView.View()
}
