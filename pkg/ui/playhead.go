package ui

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

type Playhead struct {
	Imd  *imdraw.IMDraw
	Rect pixel.Rect
	W, H float64
}

func NewPlayhead(r pixel.Rect) *Playhead {
	return &Playhead{
		Imd:  imdraw.New(nil),
		Rect: r,
		W:    r.W(),
		H:    r.H(),
	}
}

func (p *Playhead) Compose() {
	p.Imd.Clear()
	p.Imd.Color = color.RGBA{0xff, 0x42, 0x42, 0xff}
	p.Imd.Push(
		pixel.V(p.Rect.Min.X, p.Rect.Min.Y),
		pixel.V(p.Rect.Min.X, p.Rect.Max.Y),
	)
	p.Imd.Line(1)
}

func (p *Playhead) DrawTo(imd *imdraw.IMDraw) {
	p.Imd.Draw(imd)
}
