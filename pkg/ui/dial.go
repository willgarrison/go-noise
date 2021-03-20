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
	ImdStatic            *imdraw.IMDraw
	Imd                  *imdraw.IMDraw
	Rect                 pixel.Rect
	Radius               float64
	Label                string
	Value                float64
	ValueFrmt            string
	min                  float64
	max                  float64
	scale                float64
	newValue             float64
	center               pixel.Vec
	oneStepDistance      float64
	initialMousePosition pixel.Vec
	mouseInteraction     bool
	IsUnread             bool
}

// NewDial creates and returns a pointer to a Dial
func NewDial(label string, valueFrmt string, r pixel.Rect, value, min, max, scale float64) *Dial {

	d := &Dial{
		ImdStatic:       imdraw.New(nil),
		Imd:             imdraw.New(nil),
		Rect:            r,
		Radius:          r.W() / 2,
		Label:           label,
		Value:           value,
		ValueFrmt:       valueFrmt,
		newValue:        value,
		min:             min,
		max:             max,
		scale:           scale,
		center:          r.Center(),
		oneStepDistance: (math.Pi * 2) / max,
	}

	d.Compose()

	return d
}

// Compose ...
func (d *Dial) Compose() {

	d.ImdStatic.Clear()

	// Background
	d.ImdStatic.Color = color.RGBA{0xd0, 0xd0, 0xd0, 0xff}
	d.ImdStatic.Push(d.center)
	d.ImdStatic.Circle(d.Radius, 0)

	// Foreground
	d.ImdStatic.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
	d.ImdStatic.Push(d.center)
	d.ImdStatic.Circle(d.Radius-10, 0)

	d.Update()
}

// Update ...
func (d *Dial) Update() {

	xOffset := math.Sin(d.oneStepDistance * d.Value)
	yOffset := math.Cos(d.oneStepDistance * d.Value)

	// Indicator line
	d.Imd.Clear()
	d.Imd.Color = color.RGBA{0x42, 0x42, 0x42, 0xff}
	d.Imd.Push(
		pixel.V(
			d.center.X-(d.Radius-10)*xOffset,
			d.center.Y-(d.Radius-10)*yOffset,
		),
		pixel.V(
			d.center.X-d.Radius*xOffset,
			d.center.Y-d.Radius*yOffset,
		),
	)
	d.Imd.Line(3)
}

// DrawTo ...
func (d *Dial) DrawTo(imd *imdraw.IMDraw) {
	d.ImdStatic.Draw(imd)
	d.Imd.Draw(imd)
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
			d.Update()
		}
	}
}

func (d *Dial) Set(v float64) {
	d.Value = helpers.Constrain(v, d.min, d.max)
	d.newValue = d.Value
	d.Update()
}
