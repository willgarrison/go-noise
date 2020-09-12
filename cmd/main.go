package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/willgarrison/go-noise/pkg/signals"
	"github.com/willgarrison/go-noise/pkg/ui"
	"golang.org/x/image/colornames"
)

var (
	windowRect   pixel.Rect = pixel.R(0, 0, 1200, 900)
	controlsRect pixel.Rect = pixel.R(1000, 0, 200, 900)
	graphRect    pixel.Rect = pixel.R(0, 0, 1000, 900)
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

	graphChan := make(chan signals.ControlValue)
	go g.Listen(graphChan)

	for !win.Closed() {

		win.Clear(colornames.Whitesmoke)

		c.RespondToInput(win, graphChan)
		c.Draw(win)

		g.RespondToInput(win)
		g.Draw(win)

		win.Update()
	}
}
