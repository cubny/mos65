# mos65

A 6502 assembler, simulator, and step-debugger that lives in your terminal.

Type assembly, hit `Ctrl+A` to assemble, `Ctrl+R` to run. The 32×32 screen
at `$0200–$05FF` renders as half-block characters in a real TUI: snake,
breakout, and the other classic 6502asm.com demos work as you'd expect.

<img width="800" height="640" alt="mos65" src="https://github.com/user-attachments/assets/a8dbe5e0-21a9-4d6b-b2cc-874b45ab0d83" />


## Install

```sh
go install github.com/cubny/mos65@latest
```

Or build from source:

```sh
git clone https://github.com/cubny/mos65
cd mos65
go build
./mos65 testdata/snake.asm
```

Requires Go 1.24 or later. No external runtime dependencies.

## Usage

```sh
mos65 [path/to/program.asm]
```

If you pass a file, it's loaded into the editor. Otherwise you start with
a blank buffer.

### Keys

| Key       | Action                                            |
|-----------|---------------------------------------------------|
| `Tab`     | Cycle focus across editor → screen → debugger → monitor:start → monitor:length |
| `Shift+Tab` | Cycle in reverse                                |
| `Esc`     | Snap focus back to the editor                     |
| `Ctrl+A`  | Assemble the current buffer                       |
| `Ctrl+R`  | Run / stop the simulator                          |
| `Ctrl+T`  | Reset CPU and clear the screen                    |
| `Ctrl+D`  | Toggle the debugger (then `s` to step, `g` to goto) |
| `Ctrl+B`  | Toggle the memory monitor                         |
| `Ctrl+E`  | Open the buffer in `$EDITOR` (defaults to `vim`)  |
| `Ctrl+S`  | Save the buffer to its loaded path                |
| `Ctrl+F`  | Hexdump the assembled program to the messages pane |
| `Ctrl+L`  | Disassemble the assembled program                 |
| `Ctrl+N`  | Show the notes / cheat sheet                      |
| `?`       | Expand the help bar                               |
| `Ctrl+Q`  | Quit                                              |

When the **screen** pane has focus, any keystroke writes its ASCII code
to `$FF` — that's how 6502 programs read keyboard input.

### Memory map

| Address       | What                                  |
|---------------|---------------------------------------|
| `$FE`         | Fresh random byte on every instruction |
| `$FF`         | ASCII code of the last key pressed    |
| `$0200–$05FF` | 32×32 screen pixels (low nibble = colour) |
| `$0600+`      | Where assembled code lives by default  |

The 16-colour palette (low nibble of the screen byte):

```
0 black   1 white   2 red     3 cyan
4 purple  5 green   6 blue    7 yellow
8 orange  9 brown   a lt-red  b dk-gray
c gray    d lt-grn  e lt-blu  f lt-gray
```

## Credits

mos65 is a Go port of [`skilldrick/6502js`](https://github.com/skilldrick/6502js)
by Nick Morgan, itself adapted from the [6502asm.com](http://www.6502asm.com)
JavaScript simulator originally written by Stian Soreng (©2006–2010).

The 6502 instruction set, the assembler grammar, the memory layout, and
the bundled `snake.asm` all derive from those upstream projects.
See `NOTICE` for full attribution.

The TUI is built with [Charm](https://charm.sh)'s
[Bubble Tea](https://github.com/charmbracelet/bubbletea),
[Bubbles](https://github.com/charmbracelet/bubbles), and
[Lip Gloss](https://github.com/charmbracelet/lipgloss).

## License

GPL-3.0-or-later. See `LICENSE`.

Because the upstream projects are GPL-3.0, this derivative work must be
GPL-3.0 too — not a choice we get to make, but the right one for a tool
descended from a public, hackable simulator.
