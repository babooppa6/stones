package core

import (
	"fmt"
	"strings"

	"github.com/nsf/termbox-go"
)

// ListSelect displays a list of items and allows the user to select one item.
func ListSelect(title string, items []interface{}) (index int, ok bool) {
	state := TermSave()
	defer state.Restore()

	rows := []string{title}
	cols := len(title)
	for i, item := range items {
		row := fmt.Sprintf("%c) %v", i+'a', item)
		rows = append(rows, row)
		cols = Max(cols, len(row))
	}

	for y, row := range rows {
		for x, ch := range row {
			TermDraw(x, y, Glyph{ch, ColorWhite})
		}
		for x := len(row); x < cols; x++ {
			TermDraw(x, y, Glyph{' ', ColorWhite})
		}
	}
	TermRefresh()

	index = int(GetKey() - 'a')
	if index < 0 || index >= len(items) {
		return 0, false
	}
	return index, true
}

// TermTint recolors every glyph in the buffer to have the given color.
// No changes are made on screen until RefreshScreen is called.
func TermTint(c Color) {
	fg := termbox.Attribute(c)
	cells := termbox.CellBuffer()
	for i := 0; i < len(cells); i++ {
		cells[i].Fg = fg
	}
}

// Targeter allows for customization of on-screen targeting.
type Targeter struct {
	Camera  Entity
	Canvas  Entity
	Reticle Glyph
	Trace   *Glyph
	Accept  string
}

// Aim allows the user to select a target from an on-screen Camera view.
func (t Targeter) Aim() (target *Tile, ok bool) {
	state := TermSave()
	defer state.Restore()

	req := FoVRequest{}
	t.Camera.Handle(&req)
	offset := Offset{}

	var key Key
	for !strings.Contains(t.Accept, string(key)) && key != KeyEsc {
		state.Restore()

		if t.Trace != nil {
			for _, o := range Trace(offset) {
				t.Canvas.Handle(&Mark{o, *t.Trace})
			}
		}
		t.Canvas.Handle(&Mark{offset, t.Reticle})
		TermRefresh()

		key = GetKey()
		delta, ok := KeyMap[key]
		_, visible := req.FoV[offset.Add(delta)]
		if ok && visible {
			offset = offset.Add(delta)
		}
	}

	return req.FoV[offset], key != KeyEsc
}

// Aim allows the user to select a target from an on-screen Camera view.
func Aim(camera, canvas Entity, accept string) (target *Tile, ok bool) {
	return Targeter{camera, canvas, Glyph{'*', ColorRed}, nil, accept}.Aim()
}

// Mark is an Event requesting that a Glyph be drawn on Screen.
type Mark struct {
	Offset Offset
	Mark   Glyph
}

// TextDump displays a large amount of text, with scrolling.
// Useful for things like displaying large help files.
type TextDump struct {
	Title, Text string
	Fg          Color
}

// NewTextDump creates a new TextDump with the given title and text.
func NewTextDump(title, text string) *TextDump {
	return &TextDump{title, text, ColorWhite}
}

// Run displays the TextDump text, and allows the user to scroll through it.
func (t *TextDump) Run() {
	cols, rows := termbox.Size()
	lines := strings.Split(t.Text, "\n")
	currline := 0
	var key Key

	for key != KeyEsc {
		TermClear()
		for x, ch := range t.Title {
			TermDraw(x, 0, Glyph{ch, t.Fg})
		}
		for y, line := range lines[currline : currline+rows-1] {
			for x, ch := range line {
				TermDraw(x, y+1, Glyph{ch, t.Fg})
			}
		}
		TermRefresh()

		key = GetKey()
		if delta, ok := KeyMap[key]; ok && delta.X == 0 {
			currline += delta.Y
		} else if key == KeyPgup {
			currline -= cols / 2
		} else if key == KeyPgdn {
			currline += cols / 2
		}
		currline = Clamp(0, currline, len(lines)-rows)
	}
}
