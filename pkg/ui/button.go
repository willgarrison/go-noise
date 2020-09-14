package ui

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/willgarrison/go-noise/pkg/helpers"
)

// Button is an interactive UI element
type Button struct {
	Imd   *imdraw.IMDraw
	Rect  pixel.Rect
	W, H  float64
	Label string
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
}

// JustPressed ...
func (b *Button) JustPressed(pos pixel.Vec) bool {
	return helpers.PosInBounds(pos, b.Rect)
}
