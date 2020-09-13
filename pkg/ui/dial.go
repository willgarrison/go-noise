package ui

import (
	"image/color"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/willgarrison/go-noise/pkg/helpers"
)

// Dial is an interactive UI element
type Dial struct {
	Imd                              *imdraw.IMDraw
	Rect                             pixel.Rect
	W, H                             float64
	Label                            string
	Value, min, max, scale, newValue float64
	center                           pixel.Vec
	indicatorLength                  float64
	indicatorDegree                  float64
	indicatorLineWidth               float64
	initialMousePosition             pixel.Vec
	mouseInteraction                 bool
	IsUnread                         bool
}

// NewDial creates and returns a pointer to a Dial
func NewDial(label string, r pixel.Rect, value, min, max, scale float64) *Dial {

	d := &Dial{
		Imd:                imdraw.New(nil),
		Rect:               r,
		W:                  r.W(),
		H:                  r.H(),
		Label:              label,
		Value:              value,
		newValue:           value,
		min:                min,
		max:                max,
		scale:              scale,
		center:             r.Center(),
		indicatorLength:    r.W() / 2.5,
		indicatorDegree:    (math.Pi * 2) / max,
		indicatorLineWidth: 3,
	}

	d.Compose()

	return d
}

// Compose ...
func (d *Dial) Compose() {

	d.Imd.Clear()

	// Background
	d.Imd.Color = color.RGBA{0xd0, 0xd0, 0xd0, 0xff}
	d.Imd.Push(d.center)
	d.Imd.Circle(d.W/2, 0)

	// Indicator line
	d.Imd.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
	d.Imd.Push(
		d.center,
		pixel.V(
			d.center.X-d.indicatorLength*math.Sin(d.indicatorDegree*d.Value),
			d.center.Y-d.indicatorLength*math.Cos(d.indicatorDegree*d.Value),
		),
	)
	d.Imd.Line(d.indicatorLineWidth)
}

// JustPressed ...
func (d *Dial) JustPressed(pos pixel.Vec) {

	d.mouseInteraction = false

	if helpers.PosInBounds(pos, d.Rect) {
		d.mouseInteraction = true
		d.initialMousePosition = pos
	}
}

// Pressed ...
func (d *Dial) Pressed(pos pixel.Vec) {

	if d.mouseInteraction {

		if pos.Y > d.initialMousePosition.Y {
			d.newValue += (pos.Y - d.initialMousePosition.Y) * d.scale
		}
		if pos.Y < d.initialMousePosition.Y {
			d.newValue += (pos.Y - d.initialMousePosition.Y) * d.scale
		}

		d.initialMousePosition = pos

		if d.newValue != d.Value {
			d.Value = helpers.Constrain(d.newValue, d.min, d.max)
			d.newValue = d.Value
			d.IsUnread = true
			d.Compose()
		}
	}
}
