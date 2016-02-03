package core

import (
	"unicode"
)

// Label is a Visual which displays fixed text on screen.
type Label struct {
	Text string
	X, Y int
}

// Update draws the Label text at the given location.
func (l Label) Update() {
	for i, ch := range l.Text {
		TermDraw(l.X+i, l.Y, Glyph{ch, ColorWhite})
	}
}

// Border is a Visual which displays a border
type Border struct {
	Widget
	Vertical, Horizontal Glyph
}

// NewBorder creates a new Border with the given parameters.
func NewBorder(vert, horiz Glyph, x, y, w, h int) *Border {
	return &Border{Widget{x, y, w, h}, vert, horiz}
}

// Update draws the Border on screen.
func (w *Border) Update() {
	for y := 0; y < w.h; y++ {
		w.DrawRel(0, y, w.Vertical)
		w.DrawRel(w.w-1, y, w.Vertical)
	}
	for x := 0; x < w.w; x++ {
		w.DrawRel(x, 0, w.Horizontal)
		w.DrawRel(x, w.h-1, w.Horizontal)
	}
}

// TextBox is an Element which allows a user to enter custom text.
type TextBox struct {
	Text string
	Len  int
	X, Y int
}

// Update draws the current text.
func (t *TextBox) Update(selected bool) {
	var color Color
	if selected {
		color = ColorLightWhite
	} else {
		color = ColorWhite
	}

	for x := 0; x < t.Len; x++ {
		if x < len(t.Text) {
			TermDraw(t.X+x, t.Y, Glyph{rune(t.Text[x]), color})
		} else {
			TermDraw(t.X+x, t.Y, Glyph{'_', color})
		}
	}
}

// Activate lets the user enter text into the TextBox.
func (t *TextBox) Activate() FormResult {
	old := t.Text
	t.Text = ""
	t.Update(true)
	TermRefresh()

	var key Key
	for key != KeyEnter && key != KeyEsc {
		key = GetKey()
		if unicode.IsPrint(rune(key)) {
			t.Text += string(key)
		}
		t.Update(true)
		TermRefresh()
	}

	if key == KeyEsc {
		t.Text = old
	}
	return nil
}

// TODO Add TextDump (scroll through large text)
