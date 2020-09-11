package main

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/willgarrison/go-noise/pkg/signals"
	"github.com/willgarrison/go-noise/pkg/ui"
	"golang.org/x/image/colornames"
)

const (
	windowWidth    float64 = 1200.0
	windowHeight   float64 = 900.0
	graphWidth     float64 = 1000.0
	graphHeight    float64 = 900.0
	controlsWidth  float64 = 200.0
	controlsHeight float64 = 900.0
)

func main() {
	pixelgl.Run(run)
}

func run() {

	// Initialize window
	win := ui.NewWindow(windowWidth, windowHeight)

	c := ui.NewControls(1000, 0, controlsWidth, controlsHeight)
	c.Compose()

	g := ui.NewGraph(graphWidth, graphHeight)
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
