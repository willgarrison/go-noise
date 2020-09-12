package ui

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

// Playhead marks the current point on the timeline
type Playhead struct {
	Imd    *imdraw.IMDraw
	Pos    float64
	Height float64
}

// NewPlayhead creates and returns a pointer to a Playhead
func NewPlayhead(pos float64, height float64) *Playhead {

	p := &Playhead{
		Imd:    imdraw.New(nil),
		Pos:    pos,
		Height: height,
	}

	p.Compose()

	return p
}

// Compose ...
func (p *Playhead) Compose() {

	p.Imd.Clear()

	p.Imd.Color = color.RGBA{0xff, 0x42, 0x42, 0xff}
	p.Imd.Push(
		pixel.V(p.Pos, 0),
		pixel.V(p.Pos, p.Height),
	)
	p.Imd.Line(3)
}

// DrawTo ...
func (p *Playhead) DrawTo(imd *imdraw.IMDraw) {
	p.Imd.Draw(imd)
}
