package ui

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/willgarrison/go-noise/pkg/helpers"
)

type Button struct {
	Imd       *imdraw.IMDraw
	Rect      pixel.Rect
	Label     string
	isGrouped bool
	isPressed bool
	isEngaged bool
}

func NewButton(label string, r pixel.Rect) *Button {

	b := &Button{
		Imd:   imdraw.New(nil),
		Rect:  r,
		Label: label,
	}

	b.Compose()

	return b
}

func (b *Button) Compose() {

	b.Imd.Clear()

	b.Imd.Color = color.RGBA{0x00, 0x00, 0x00, 0xff}
	b.Imd.Push(b.Rect.Min, b.Rect.Max)
	b.Imd.Rectangle(1)

	if b.isPressed {
		b.Imd.Color = color.RGBA{0xdd, 0xdd, 0xdd, 0xff}
		b.Imd.Push(b.Rect.Min, b.Rect.Max)
		b.Imd.Rectangle(0)
	}

	if b.isGrouped {
		b.Imd.Color = color.RGBA{0xee, 0xee, 0xee, 0xff}
		if b.isEngaged {
			b.Imd.Color = color.RGBA{0x36, 0xaf, 0xcf, 0xff}
		}
		b.Imd.Push(pixel.V(b.Rect.Min.X+10, b.Rect.Min.Y+10), pixel.V(b.Rect.Min.X+20, b.Rect.Min.Y+20))
		b.Imd.Rectangle(0)
	}
}

func (b *Button) PosInBounds(pos pixel.Vec) bool {
	return helpers.PosInBounds(pos, b.Rect)
}

func (b *Button) SetGrouped(state bool) {
	b.isGrouped = state
	b.Compose()
}

func (b *Button) SetPressed(state bool) {
	b.isPressed = state
	b.Compose()
}

func (b *Button) SetEngaged(state bool) {
	b.isEngaged = state
	b.Compose()
}
