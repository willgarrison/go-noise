package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/willgarrison/go-noise/pkg/signals"
	"github.com/willgarrison/go-noise/pkg/ui"
	"golang.org/x/image/colornames"
)

var (
	windowRect   pixel.Rect = pixel.R(0, 0, 1200, 900)
	controlsRect pixel.Rect = pixel.R(1000, 0, 200, 900)
	graphRect    pixel.Rect = pixel.R(0, 0, 1000, 900)
	bpm          float64    = 60
)

func main() {
	pixelgl.Run(run)
}

func run() {

	// Initialize window
	win := ui.NewWindow(windowRect.W(), windowRect.H())

	c := ui.NewControls(controlsRect)
	c.Compose()

	g := ui.NewGraph(graphRect)
	g.Compose()

	ctrlChan := make(chan signals.ControlValue)
	go g.Listen(ctrlChan)

	p := ui.NewPlayhead(graphRect.Min.X, graphRect.Max.Y)
	p.Compose()

	imdBatch := imdraw.New(nil)

	for !win.Closed() {

		win.Clear(colornames.Whitesmoke)

		imdBatch.Clear()

		c.RespondToInput(win, ctrlChan)
		c.DrawTo(imdBatch)
		c.Typ.TxtBatch.Draw(win)

		g.RespondToInput(win)
		g.DrawTo(imdBatch)

		p.DrawTo(imdBatch)

		imdBatch.Draw(win)

		win.Update()
	}
}
