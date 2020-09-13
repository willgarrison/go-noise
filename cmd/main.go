package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/willgarrison/go-noise/pkg/metronome"
	"github.com/willgarrison/go-noise/pkg/midi"
	"github.com/willgarrison/go-noise/pkg/ui"
	"golang.org/x/image/colornames"
)

var (
	windowRect   pixel.Rect = pixel.R(0, 0, 1200, 900)
	controlsRect pixel.Rect = pixel.R(1000, 0, 200, 900)
	graphRect    pixel.Rect = pixel.R(20, 20, 980, 880)
)

func main() {
	pixelgl.Run(run)
}

func run() {

	// Initialize midi output
	audio, err := midi.New()
	if err != nil {
		panic(err.Error())
	}
	defer audio.Driver.Close()

	// Initialize window
	win := ui.NewWindow(windowRect.W(), windowRect.H())

	// Initialize batch
	imdBatch := imdraw.New(nil)

	// Initialize graph
	g := ui.NewGraph(graphRect, audio.Output)
	g.Compose()

	// Initialize metronome
	m := metronome.New(g.Bpm)
	m.AddBeatChannel(g.BeatChannel)

	// Initialize controls
	c := ui.NewControls(controlsRect)
	c.AddCtrlChannel(g.CtrlChannel)
	c.AddCtrlChannel(m.CtrlChannel)
	c.Compose()

	// Start metronome
	m.Start()

	for !win.Closed() {

		win.Clear(colornames.Whitesmoke)

		imdBatch.Clear()

		c.RespondToInput(win)
		c.DrawTo(imdBatch)
		c.Typ.TxtBatch.Draw(win)

		g.RespondToInput(win)
		g.DrawTo(imdBatch)

		imdBatch.Draw(win)

		win.Update()
	}
}
