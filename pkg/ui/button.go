package ui

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/willgarrison/go-noise/pkg/helpers"
)

// Button is an interactive UI element
type Button struct {
	Imd     *imdraw.IMDraw
	Rect    pixel.Rect
	W, H    float64
	Label   string
	Grouped bool
	Active  bool
	Pressed bool
}

// NewButton creates and returns a pointer to a Button
func NewButton(label string, r pixel.Rect) *Button {

	b := &Button{
		Imd:   imdraw.New(nil),
		Rect:  r,
		W:     r.W(),
		H:     r.H(),
		Label: label,
	}

	b.Compose()

	return b
}

// Compose ...
func (b *Button) Compose() {

	b.Imd.Clear()

	b.Imd.Color = color.RGBA{0x00, 0x00, 0x00, 0xff}
	b.Imd.Push(b.Rect.Min, b.Rect.Max)
	b.Imd.Rectangle(1)

	if b.Pressed {
		b.Imd.Color = color.RGBA{0xdd, 0xdd, 0xdd, 0xff}
		b.Imd.Push(b.Rect.Min, b.Rect.Max)
		b.Imd.Rectangle(0)
	}

	if b.Grouped {
		b.Imd.Color = color.RGBA{0xee, 0xee, 0xee, 0xff}
		if b.Active {
			b.Imd.Color = color.RGBA{0x36, 0xaf, 0xcf, 0xff}
		}
		b.Imd.Push(pixel.V(b.Rect.Min.X+10, b.Rect.Min.Y+10), pixel.V(b.Rect.Min.X+20, b.Rect.Min.Y+20))
		b.Imd.Rectangle(0)
	}
}

// PosInBounds ...
func (b *Button) PosInBounds(pos pixel.Vec) bool {
	return helpers.PosInBounds(pos, b.Rect)
}

// SetGrouped ...
func (b *Button) SetGrouped(state bool) {
	b.Grouped = state
	b.Compose()
}

// SetActive ...
func (b *Button) SetActive(state bool) {
	b.Active = state
	b.Compose()
}

// SetPressed ...
func (b *Button) SetPressed(state bool) {
	b.Pressed = state
	b.Compose()
}
