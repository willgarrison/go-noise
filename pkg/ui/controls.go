package ui

import (
	"fmt"
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/willgarrison/go-noise/pkg/helpers"
	"github.com/willgarrison/go-noise/pkg/signals"
)

// Controls ...
type Controls struct {
	X, Y, W, H float64
	Dials      []*Dial
	Buttons    []*Button
	Imd        *imdraw.IMDraw
	ImdBatch   *imdraw.IMDraw
	typ        *Typography
}

// NewControls ...
func NewControls(x, y, w, h float64) *Controls {

	c := new(Controls)

	c.X = x
	c.Y = y
	c.W = w
	c.H = h

	c.ResetButtons()
	c.ResetDials()

	c.Imd = imdraw.New(nil)
	c.ImdBatch = imdraw.New(nil)

	c.typ = NewTypography()

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
	c.Dials = make([]*Dial, 7)
	c.Dials[0] = NewDial("frequency", c.X+20, c.Y+160, dialSize, 0.3, 0.01, 3.0, 0.001)
	c.Dials[1] = NewDial("lacunarity", c.X+110, c.Y+160, dialSize, 0.5, 0.01, 3.0, 0.01)
	c.Dials[2] = NewDial("gain", c.X+20, c.Y+270, dialSize, 2.0, 0.01, 3.0, 0.1)
	c.Dials[3] = NewDial("octaves", c.X+110, c.Y+270, dialSize, 5, 1, 10, 1)
	c.Dials[4] = NewDial("xSteps", c.X+20, c.Y+380, dialSize, 32, 4, 64, 1)
	c.Dials[5] = NewDial("ySteps", c.X+110, c.Y+380, dialSize, 48, 4, 96, 1)
	c.Dials[6] = NewDial("offset", c.X+20, c.Y+490, dialSize, 500, 0, 1000, 1)
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

	c.typ.txtBatch.Clear()

	for i := range c.Buttons {

		// Labels
		str := c.Buttons[i].Label
		strX := c.Buttons[i].X + (c.Buttons[i].W / 2) - (c.typ.txt.BoundsOf(str).W() / 2)
		strY := c.Buttons[i].Y + (c.Buttons[i].H / 2) - (c.typ.txt.BoundsOf(str).H() / 2)
		helpers.DrawTextToBatch(str, pixel.V(strX, strY), c.typ.txtBatch, c.typ.txt)
	}

	for i := range c.Dials {

		// Labels
		str := c.Dials[i].Label
		strX := c.Dials[i].center.X - (c.typ.txt.BoundsOf(str).W() / 2)
		strY := c.Dials[i].y - 20
		helpers.DrawTextToBatch(str, pixel.V(strX, strY), c.typ.txtBatch, c.typ.txt)

		// Values
		str = fmt.Sprintf("%.3f", c.Dials[i].Value)
		strX = c.Dials[i].center.X - (c.typ.txt.BoundsOf(str).W() / 2)
		strY = c.Dials[i].y - 10
		helpers.DrawTextToBatch(str, pixel.V(strX, strY), c.typ.txtBatch, c.typ.txt)
	}
}

// Draw ...
func (c *Controls) Draw(win *pixelgl.Window) {

	// Draw static content
	c.Imd.Draw(win)

	// Draw dynamic content to batch
	c.ImdBatch.Clear()
	for i := range c.Buttons {
		c.Buttons[i].Imd.Draw(c.ImdBatch)
	}
	for i := range c.Dials {
		c.Dials[i].Imd.Draw(c.ImdBatch)
	}

	// Draw batch
	c.ImdBatch.Draw(win)

	// Draw textBatch
	c.typ.txtBatch.Draw(win)
}

// RespondToInput ...
func (c *Controls) RespondToInput(win *pixelgl.Window, sendToChan chan signals.ControlValue) {

	// Listen for mouse input
	if win.JustPressed(pixelgl.MouseButtonLeft) {

		pos := win.MousePosition()

		for i := range c.Buttons {

			// Reset button
			if c.Buttons[i].JustPressed(pos) {
				cv := signals.ControlValue{
					Label: c.Buttons[i].Label,
					Value: 1.0,
				}
				c.Send(sendToChan, cv)
				c.ResetDials()
			}
		}

		for i := range c.Dials {
			c.Dials[i].JustPressed(pos)
		}
	}

	if win.Pressed(pixelgl.MouseButtonLeft) {

		pos := win.MousePosition()

		for i := range c.Dials {

			c.Dials[i].Pressed(pos)

			if c.Dials[i].IsUnread {

				cv := signals.ControlValue{
					Label: c.Dials[i].Label,
					Value: c.Dials[i].Value,
				}

				c.Send(sendToChan, cv)
				c.Dials[i].IsUnread = false
				c.Compose()
			}
		}
	}
}

// Send ...
func (c *Controls) Send(sendToChan chan signals.ControlValue, newControlValue signals.ControlValue) {
	sendToChan <- newControlValue
}
