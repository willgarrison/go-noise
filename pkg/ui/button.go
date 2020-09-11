package ui

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/willgarrison/go-noise/pkg/helpers"
)

// Button is an interactive UI element
type Button struct {
	Imd        *imdraw.IMDraw
	Bounds     []pixel.Vec
	Label      string
	X, Y, W, H float64
}

// NewButton creates and returns a pointer to a Button
func NewButton(label string, x, y, w, h float64) *Button {

	b := &Button{
		Imd: imdraw.New(nil),
		Bounds: []pixel.Vec{
			pixel.V(x, y),
			pixel.V(x+w, y+h),
		},
		Label: label,
		X:     x,
		Y:     y,
		W:     w,
		H:     h,
	}

	b.Compose()

	return b
}

// Compose ...
func (b *Button) Compose() {

	b.Imd.Clear()

	b.Imd.Color = color.RGBA{0x42, 0x42, 0x42, 0xff}
	b.Imd.Push(b.Bounds[0], b.Bounds[1])
	b.Imd.Rectangle(1)
}

// JustPressed ...
func (b *Button) JustPressed(pos pixel.Vec) bool {
	return helpers.PosInBounds(pos, b.Bounds)
}
