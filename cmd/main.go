package main

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/willgarrison/go-noise/pkg/ui"
	"golang.org/x/image/colornames"
)

const (
	width  float64 = 800.0
	height float64 = 600.0
)

func main() {
	pixelgl.Run(run)
}

func run() {

	// Initialize window
	win := ui.NewWindow(width, height)

	g := ui.NewGraph(width, height)

	g.Generate()

	for !win.Closed() {

		win.Clear(colornames.Whitesmoke)

		g.RespondToInput(win)

		g.Draw(win)

		win.Update()
	}
}
