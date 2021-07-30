package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/willgarrison/go-noise/pkg/metronome"
	"github.com/willgarrison/go-noise/pkg/midi"
	"github.com/willgarrison/go-noise/pkg/session"
	"github.com/willgarrison/go-noise/pkg/ui"
	"golang.org/x/image/colornames"
)

var (
	windowRect   pixel.Rect = pixel.R(0, 0, 1200, 960)
	graphRect    pixel.Rect = pixel.R(80.01, 60.01, 980, 940)
	controlsRect pixel.Rect = pixel.R(1000, 0, 1200, 960)
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

	// Initialize session
	s := session.NewSession()

	// Initialize graph
	g := ui.NewGraph(graphRect, audio.Output, &s.SessionData)
	g.Compose()

	// Initialize metronome
	m := metronome.New(&s.SessionData)

	// Initialize controls
	c := ui.NewControls(controlsRect, &s.SessionData)
	c.Compose()

	// Connect session outputs
	s.AddOutputChannel(c.InputSessionChannel)
	s.AddOutputChannel(g.InputSessionChannel)
	s.AddOutputChannel(m.InputSessionChannel)

	// Connect metronome outputs
	m.AddOutputChannel(g.InputBeatChannel)

	// Connect control outputs
	c.AddOutputChannel(s.InputCtrlChannel)
	c.AddOutputChannel(g.InputCtrlChannel)
	c.AddOutputChannel(m.InputCtrlChannel)

	// Start metronome
	m.Start()

	for !win.Closed() {

		win.Clear(colornames.Whitesmoke)

		imdBatch.Clear()

		c.RespondToInput(win)
		c.DrawTo(imdBatch)

		g.RespondToInput(win)
		g.DrawTo(imdBatch)

		imdBatch.Draw(win)
		c.Typ.TxtBatch.Draw(win)
		g.Typ.TxtBatch.Draw(win)

		win.Update()
	}
}
