package ui

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

// Playhead marks the current point on the timeline
type Playhead struct {
	Imd  *imdraw.IMDraw
	Rect pixel.Rect
	W, H float64
}

// NewPlayhead creates and returns a pointer to a Playhead
func NewPlayhead(r pixel.Rect) *Playhead {

	p := &Playhead{
		Imd:  imdraw.New(nil),
		Rect: r,
		W:    r.W(),
		H:    r.H(),
	}

	return p
}

// Compose ...
func (p *Playhead) Compose() {

	p.Imd.Clear()

	p.Imd.Color = color.RGBA{0xff, 0x42, 0x42, 0xff}
	p.Imd.Push(
		pixel.V(p.Rect.Min.X, p.Rect.Min.Y),
		pixel.V(p.Rect.Min.X, p.Rect.Max.Y),
	)
	p.Imd.Line(1)
}

// DrawTo ...
func (p *Playhead) DrawTo(imd *imdraw.IMDraw) {
	p.Imd.Draw(imd)
}
