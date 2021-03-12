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
	Rect         pixel.Rect
	W, H         float64
	Dials        []*Dial
	Buttons      []*Button
	ModeButtons  []*Button
	Imd          *imdraw.IMDraw
	ImdBatch     *imdraw.IMDraw
	Typ          *Typography
	CtrlChannels []chan signals.CtrlSignal
}

// NewControls ...
func NewControls(r pixel.Rect) *Controls {

	c := new(Controls)

	c.Rect = r
	c.W = r.W()
	c.H = r.H()

	c.ResetButtons()
	c.ResetDials()

	c.Imd = imdraw.New(nil)
	c.ImdBatch = imdraw.New(nil)

	c.Typ = NewTypography()

	return c
}

// ResetButtons ..
func (c *Controls) ResetButtons() {

	buttonWidths := []float64{
		75.0,
		160.0,
	}
	buttonHeights := []float64{
		30.0,
		60.0,
	}

	columnPos := []float64{
		c.Rect.Min.X + 20,
		c.Rect.Min.X + 105,
	}

	rowPos := []float64{
		c.Rect.Min.Y + 580.01,
		c.Rect.Min.Y + 620.01,
		c.Rect.Min.Y + 660.01,
		c.Rect.Min.Y + 700.01,
		c.Rect.Min.Y + 740.01,
		c.Rect.Min.Y + 800.01,
		c.Rect.Min.Y + 880.01,
	}

	c.ModeButtons = []*Button{
		NewButton("major", pixel.R(columnPos[0], rowPos[0], columnPos[0]+buttonWidths[1], rowPos[0]+buttonHeights[0])),
		NewButton("natural", pixel.R(columnPos[0], rowPos[1], columnPos[0]+buttonWidths[1], rowPos[1]+buttonHeights[0])),
		NewButton("harmonic", pixel.R(columnPos[0], rowPos[2], columnPos[0]+buttonWidths[1], rowPos[2]+buttonHeights[0])),
		NewButton("melodic", pixel.R(columnPos[0], rowPos[3], columnPos[0]+buttonWidths[1], rowPos[3]+buttonHeights[0])),
		NewButton("pentatonic", pixel.R(columnPos[0], rowPos[4], columnPos[0]+buttonWidths[1], rowPos[4]+buttonHeights[0])),
	}

	c.Buttons = []*Button{
		NewButton("reset", pixel.R(columnPos[0], c.Rect.Min.Y+20, c.Rect.Max.X-20, c.Rect.Min.Y+90)),
		NewButton("play", pixel.R(columnPos[0], rowPos[5], columnPos[0]+buttonWidths[0], rowPos[5]+buttonHeights[1])),
		NewButton("stop", pixel.R(columnPos[1], rowPos[5], columnPos[1]+buttonWidths[0], rowPos[5]+buttonHeights[1])),
		NewButton("save", pixel.R(columnPos[0], rowPos[6], columnPos[0]+buttonWidths[0], rowPos[6]+buttonHeights[1])),
		NewButton("load", pixel.R(columnPos[1], rowPos[6], columnPos[1]+buttonWidths[0], rowPos[6]+buttonHeights[1])),
	}

	for i := range c.ModeButtons {
		c.ModeButtons[i].SetGrouped(true)
	}

	c.ModeButtons[0].SetActive(true)
}

// ResetDials ..
func (c *Controls) ResetDials() {

	dialWidth := 70.0
	dialHeight := 70.0

	columnPos := []float64{
		c.Rect.Min.X + 20,
		c.Rect.Min.X + 110,
	}

	rowPos := []float64{
		c.Rect.Min.Y + 130,
		c.Rect.Min.Y + 220,
		c.Rect.Min.Y + 310,
		c.Rect.Min.Y + 400,
		c.Rect.Min.Y + 490,
	}

	c.Dials = make([]*Dial, 9)
	c.Dials[0] = NewDial("freq", "%.3f", pixel.R(columnPos[0], rowPos[0], columnPos[0]+dialWidth, rowPos[0]+dialHeight), 0.3, 0.01, 3.0, 0.001)
	c.Dials[1] = NewDial("space", "%.2f", pixel.R(columnPos[1], rowPos[0], columnPos[1]+dialWidth, rowPos[0]+dialHeight), 0.9, 0.01, 3.0, 0.01)
	c.Dials[2] = NewDial("gain", "%.1f", pixel.R(columnPos[0], rowPos[1], columnPos[0]+dialWidth, rowPos[1]+dialHeight), 2.0, 0.01, 3.0, 0.1)
	c.Dials[3] = NewDial("octs", "%.1f", pixel.R(columnPos[1], rowPos[1], columnPos[1]+dialWidth, rowPos[1]+dialHeight), 5, 1, 10, 1)
	c.Dials[4] = NewDial("x", "%.1f", pixel.R(columnPos[0], rowPos[2], columnPos[0]+dialWidth, rowPos[2]+dialHeight), 16, 4, 64, 1)
	c.Dials[5] = NewDial("y", "%.1f", pixel.R(columnPos[1], rowPos[2], columnPos[1]+dialWidth, rowPos[2]+dialHeight), 24, 4, 48, 1)
	c.Dials[6] = NewDial("pos", "%.1f", pixel.R(columnPos[0], rowPos[3], columnPos[0]+dialWidth, rowPos[3]+dialHeight), 0, 0, 1000, 1)
	c.Dials[7] = NewDial("bpm", "%.1f", pixel.R(columnPos[1], rowPos[3], columnPos[1]+dialWidth, rowPos[3]+dialHeight), 180, 1, 960, 1)
	c.Dials[8] = NewDial("rel", "%.1f", pixel.R(columnPos[1], rowPos[4], columnPos[1]+dialWidth, rowPos[4]+dialHeight), 1, 0, 8, 1)
}

// Compose ...
func (c *Controls) Compose() {

	c.Imd.Color = color.RGBA{0x00, 0x00, 0x00, 0xff}

	// Left
	c.Imd.Push(
		pixel.V(c.Rect.Min.X, c.Rect.Min.Y),
		pixel.V(c.Rect.Min.X, c.Rect.Max.Y),
	)
	c.Imd.Line(1)

	c.Typ.TxtBatch.Clear()

	for i := range c.Buttons {

		// Labels
		str := c.Buttons[i].Label
		strX := c.Buttons[i].Rect.Min.X + (c.Buttons[i].W / 2) - (c.Typ.Txt.BoundsOf(str).W() / 2)
		strY := c.Buttons[i].Rect.Min.Y + (c.Buttons[i].H / 2) - (c.Typ.Txt.BoundsOf(str).H() / 3)
		c.Typ.DrawTextToBatch(str, pixel.V(strX, strY), color.RGBA{0x00, 0x00, 0x00, 0xff}, c.Typ.TxtBatch, c.Typ.Txt)
	}

	for i := range c.ModeButtons {

		// Labels
		str := c.ModeButtons[i].Label
		strX := c.ModeButtons[i].Rect.Min.X + (c.ModeButtons[i].W / 2) - (c.Typ.Txt.BoundsOf(str).W() / 2)
		strY := c.ModeButtons[i].Rect.Min.Y + (c.ModeButtons[i].H / 2) - (c.Typ.Txt.BoundsOf(str).H() / 3)
		c.Typ.DrawTextToBatch(str, pixel.V(strX, strY), color.RGBA{0x00, 0x00, 0x00, 0xff}, c.Typ.TxtBatch, c.Typ.Txt)
	}

	for i := range c.Dials {

		// Values
		str := fmt.Sprintf(c.Dials[i].ValueFrmt, c.Dials[i].Value)
		strX := c.Dials[i].center.X - (c.Typ.Txt.BoundsOf(str).W() / 2)
		strY := c.Dials[i].center.Y - (c.Typ.Txt.BoundsOf(str).H() / 3) + 5
		c.Typ.DrawTextToBatch(str, pixel.V(strX, strY), color.RGBA{0x00, 0x00, 0x00, 0xff}, c.Typ.TxtBatch, c.Typ.Txt)

		// Labels
		str = c.Dials[i].Label
		strX = c.Dials[i].center.X - (c.Typ.Txt.BoundsOf(str).W() / 2)
		strY = c.Dials[i].center.Y - (c.Typ.Txt.BoundsOf(str).H() / 3) - 10
		c.Typ.DrawTextToBatch(str, pixel.V(strX, strY), color.RGBA{0x42, 0x42, 0x42, 0xff}, c.Typ.TxtBatch, c.Typ.Txt)
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
	for i := range c.ModeButtons {
		c.ModeButtons[i].Imd.Draw(c.ImdBatch)
	}
	for i := range c.Dials {
		c.Dials[i].DrawTo(c.ImdBatch)
	}

	// Draw batch
	c.ImdBatch.Draw(imd)
}

// RespondToInput ...
func (c *Controls) RespondToInput(win *pixelgl.Window) {

	// Key commands:
	// Play/Pause with Spacebar
	if win.JustPressed(pixelgl.KeySpace) {
		ctrlSignal := signals.CtrlSignal{
			Label: "toggle",
			Value: 1.0,
		}
		c.Send(ctrlSignal)
	}

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
				if c.Buttons[i].Label == "reset" {
					c.ResetDials()
				}
				c.Send(ctrlSignal)
				c.Compose()
			}
		}

		// ModeButtons
		// Check if a modeButton was pressed
		modeButtonPressedIndex := -1
		for i := range c.ModeButtons {
			if c.ModeButtons[i].JustPressed(pos) {
				modeButtonPressedIndex = i
			}
		}
		// If a modeButton was pressed
		if modeButtonPressedIndex > -1 {
			// Deactivate all mode buttons
			for i := range c.ModeButtons {
				c.ModeButtons[i].SetActive(false)
			}
			// Activate the pressed mode button
			c.ModeButtons[modeButtonPressedIndex].SetActive(true)
			// Send signal
			ctrlSignal := signals.CtrlSignal{
				Label: c.ModeButtons[modeButtonPressedIndex].Label,
				Value: 1.0,
			}
			c.Send(ctrlSignal)
			c.Compose()
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
