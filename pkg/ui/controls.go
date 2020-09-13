package ui

import (
	"fmt"
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/willgarrison/go-noise/pkg/signals"
)

// Controls ...
type Controls struct {
	X, Y, W, H   float64
	Dials        []*Dial
	Buttons      []*Button
	Imd          *imdraw.IMDraw
	ImdBatch     *imdraw.IMDraw
	Typ          *Typography
	CtrlChannels []chan signals.CtrlSignal
}

// NewControls ...
func NewControls(r pixel.Rect) *Controls {

	c := new(Controls)

	c.X = r.Min.X
	c.Y = r.Min.Y
	c.W = r.Max.X
	c.H = r.Max.Y

	c.ResetButtons()
	c.ResetDials()

	c.Imd = imdraw.New(nil)
	c.ImdBatch = imdraw.New(nil)

	c.Typ = NewTypography()

	return c
}

// ResetButtons ..
func (c *Controls) ResetButtons() {

	c.Buttons = make([]*Button, 1)
	c.Buttons[0] = NewButton("reset", c.X+20, c.Y+20, 160, 90)
}

// ResetDials ..
func (c *Controls) ResetDials() {

	dialSize := 75.0
	c.Dials = make([]*Dial, 8)
	c.Dials[0] = NewDial("frequency", c.X+20, c.Y+160, dialSize, 0.3, 0.01, 3.0, 0.001)
	c.Dials[1] = NewDial("lacunarity", c.X+110, c.Y+160, dialSize, 0.9, 0.01, 3.0, 0.01)
	c.Dials[2] = NewDial("gain", c.X+20, c.Y+270, dialSize, 2.0, 0.01, 3.0, 0.1)
	c.Dials[3] = NewDial("octaves", c.X+110, c.Y+270, dialSize, 5, 1, 10, 1)
	c.Dials[4] = NewDial("xSteps", c.X+20, c.Y+380, dialSize, 8, 4, 64, 1)
	c.Dials[5] = NewDial("ySteps", c.X+110, c.Y+380, dialSize, 24, 4, 48, 1)
	c.Dials[6] = NewDial("offset", c.X+20, c.Y+490, dialSize, 500, 0, 1000, 1)
	c.Dials[7] = NewDial("bpm", c.X+110, c.Y+490, dialSize, 120, 1, 960, 1)
}

// Compose ...
func (c *Controls) Compose() {

	// Indicator line
	c.Imd.Color = color.RGBA{0x00, 0x00, 0x00, 0xff}
	c.Imd.Push(
		pixel.V(c.X+1, c.Y),
		pixel.V(c.X+1, c.Y+c.H),
	)
	c.Imd.Line(1)

	c.Typ.TxtBatch.Clear()

	for i := range c.Buttons {

		// Labels
		str := c.Buttons[i].Label
		strX := c.Buttons[i].X + (c.Buttons[i].W / 2) - (c.Typ.Txt.BoundsOf(str).W() / 2)
		strY := c.Buttons[i].Y + (c.Buttons[i].H / 2) - (c.Typ.Txt.BoundsOf(str).H() / 2)
		c.Typ.DrawTextToBatch(str, pixel.V(strX, strY), c.Typ.TxtBatch, c.Typ.Txt)
	}

	for i := range c.Dials {

		// Labels
		str := c.Dials[i].Label
		strX := c.Dials[i].center.X - (c.Typ.Txt.BoundsOf(str).W() / 2)
		strY := c.Dials[i].y - 20
		c.Typ.DrawTextToBatch(str, pixel.V(strX, strY), c.Typ.TxtBatch, c.Typ.Txt)

		// Values
		str = fmt.Sprintf("%.3f", c.Dials[i].Value)
		strX = c.Dials[i].center.X - (c.Typ.Txt.BoundsOf(str).W() / 2)
		strY = c.Dials[i].y - 10
		c.Typ.DrawTextToBatch(str, pixel.V(strX, strY), c.Typ.TxtBatch, c.Typ.Txt)
	}
}

// DrawTo ...
func (c *Controls) DrawTo(imd *imdraw.IMDraw) {

	// Draw static content
	c.Imd.Draw(imd)

	// Draw dynamic content to batch
	c.ImdBatch.Clear()
	for i := range c.Buttons {
		c.Buttons[i].Imd.Draw(c.ImdBatch)
	}
	for i := range c.Dials {
		c.Dials[i].Imd.Draw(c.ImdBatch)
	}

	// Draw batch
	c.ImdBatch.Draw(imd)
}

// RespondToInput ...
func (c *Controls) RespondToInput(win *pixelgl.Window) {

	if win.JustPressed(pixelgl.MouseButtonLeft) {

		pos := win.MousePosition()

		// Buttons
		for i := range c.Buttons {

			// Reset button
			if c.Buttons[i].JustPressed(pos) {
				ctrlSignal := signals.CtrlSignal{
					Label: c.Buttons[i].Label,
					Value: 1.0,
				}
				c.Send(ctrlSignal)
				c.ResetDials()
				c.Compose()
			}
		}

		// Dails
		for i := range c.Dials {
			c.Dials[i].JustPressed(pos)
		}
	}

	if win.Pressed(pixelgl.MouseButtonLeft) {

		pos := win.MousePosition()

		// Dails
		for i := range c.Dials {

			c.Dials[i].Pressed(pos)

			if c.Dials[i].IsUnread {
				ctrlSignal := signals.CtrlSignal{
					Label: c.Dials[i].Label,
					Value: c.Dials[i].Value,
				}
				c.Send(ctrlSignal)
				c.Dials[i].IsUnread = false
				c.Compose()
			}
		}
	}
}

// AddCtrlChannel ...
func (c *Controls) AddCtrlChannel(ctrlChannel chan signals.CtrlSignal) {
	c.CtrlChannels = append(c.CtrlChannels, ctrlChannel)
}

// Send ...
func (c *Controls) Send(ctrlSignal signals.CtrlSignal) {
	// Send ctrl signal to all subscribers
	for index := range c.CtrlChannels {
		c.CtrlChannels[index] <- ctrlSignal
	}
}
