package ui

import (
	"image/color"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/willgarrison/go-noise/pkg/helpers"
	"github.com/willgarrison/go-noise/pkg/signals"
	"github.com/willgarrison/go-noise/pkg/simplexnoise"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/writer"
)

// Graph ...
type Graph struct {
	X, Y, W, H     float64
	Matrix         [][]uint32
	Imd            *imdraw.IMDraw
	CtrlChannel    chan signals.CtrlSignal
	BeatChannel    chan signals.BeatSignal
	BeatIndex      uint8
	NotesOn        []uint8
	Scale          []uint8
	MidiWriter     *writer.Writer
	MidiOutput     midi.Out
	Playhead       *Playhead
	ShouldReset    bool
	SignalReceived bool
	Frequency      float32
	Lacunarity     float32
	Gain           float32
	Octaves        uint8
	XSteps         uint32
	YSteps         uint32
	Offset         uint32
}

// NewGraph ...
func NewGraph(r pixel.Rect, ao midi.Out) *Graph {

	g := new(Graph)

	g.X = r.Min.X
	g.Y = r.Min.Y
	g.W = r.Max.X
	g.H = r.Max.Y

	g.Imd = imdraw.New(nil)

	// Initialize playhead
	g.Playhead = NewPlayhead(pixel.R(g.X, g.Y, 3, g.H))
	g.Playhead.Compose()

	g.Reset()

	g.Scale = []uint8{
		0, 2, 4, 5, 7, 9, 11,
		12, 14, 16, 17, 19, 21, 23,
		24, 26, 28, 29, 31, 33, 35,
		36, 38, 40, 41, 43, 45, 47,
		48, 50, 52, 53, 55, 57, 59,
		60, 62, 64, 65, 67, 69, 71,
		72, 74, 76, 77, 79, 81, 83,
		84, 86, 88, 89, 91, 93, 95,
		96, 98, 100, 101, 103, 105, 107,
	}

	g.MidiWriter = writer.New(ao)
	g.MidiWriter.SetChannel(1)

	g.CtrlChannel = make(chan signals.CtrlSignal)
	g.ListenToCtrlChannel()

	g.BeatChannel = make(chan signals.BeatSignal)
	g.ListenToBeatChannel()

	return g
}

// Reset ...
func (g *Graph) Reset() {
	g.Frequency = 0.3
	g.Lacunarity = 0.9
	g.Gain = 2.0
	g.Octaves = 5
	g.XSteps = 8
	g.YSteps = 24
	g.Offset = 500
}

// Compose ...
func (g *Graph) Compose() {

	// Reset Matrix
	g.Matrix = make([][]uint32, int(g.XSteps))
	for i := range g.Matrix {
		g.Matrix[i] = make([]uint32, int(g.YSteps))
	}

	xPos := uint32(0)
	for xPos < g.XSteps {
		val := simplexnoise.Fbm(float32(xPos+g.Offset), 0, g.Frequency, g.Lacunarity, g.Gain, int(g.Octaves))
		yPos := uint32(math.Round(helpers.ReRange(float64(val), -1, 1, 0, float64(g.YSteps-1))))
		g.Matrix[xPos][yPos] = 1
		xPos++
	}

	g.Imd.Clear()

	// Draw active blocks
	g.Imd.Color = color.RGBA{0x00, 0x00, 0x00, 0xff}
	blockWidth := g.W / float64(g.XSteps)
	blockHeight := g.H / float64(g.YSteps)
	for x := range g.Matrix {
		for y := range g.Matrix[x] {
			if g.Matrix[x][y] == 1 {
				x1 := g.X + (float64(x) * g.W / float64(g.XSteps))
				y1 := g.Y + (float64(y) * g.H / float64(g.YSteps))
				x2 := float64(x)*g.W/float64(g.XSteps) + blockWidth
				y2 := float64(y)*g.H/float64(g.YSteps) + blockHeight
				g.Imd.Push(
					pixel.V(x1, y1),
					pixel.V(x2, y2),
				)
				g.Imd.Rectangle(0)
			}
		}
	}
}

// DrawTo ...
func (g *Graph) DrawTo(imd *imdraw.IMDraw) {
	g.Imd.Draw(imd)
	g.Playhead.DrawTo(imd)
}

// RespondToInput ...
func (g *Graph) RespondToInput(win *pixelgl.Window) {
	if g.SignalReceived {
		g.SignalReceived = false
		g.Compose()
	}
}

// ListenToCtrlChannel ...
func (g *Graph) ListenToCtrlChannel() {
	go func() {
		for {
			select {
			case ctrlSignal := <-g.CtrlChannel:
				switch ctrlSignal.Label {
				case "reset":
					g.Reset()
				case "frequency":
					g.Frequency = float32(ctrlSignal.Value)
				case "lacunarity":
					g.Lacunarity = float32(ctrlSignal.Value)
				case "gain":
					g.Gain = float32(ctrlSignal.Value)
				case "octaves":
					g.Octaves = uint8(ctrlSignal.Value)
				case "xSteps":
					g.XSteps = uint32(ctrlSignal.Value)
				case "ySteps":
					g.YSteps = uint32(ctrlSignal.Value)
				case "offset":
					g.Offset = uint32(ctrlSignal.Value)
				}
				g.SignalReceived = true
			}
		}
	}()
}

// ListenToBeatChannel ...
func (g *Graph) ListenToBeatChannel() {
	go func() {
		for {
			select {
			case beatSignal := <-g.BeatChannel:
				g.BeatIndex = beatSignal.Value % uint8(len(g.Matrix))
				g.Stop()
				for y, val := range g.Matrix[g.BeatIndex] {
					if val == 1 {
						g.NotesOn = append(g.NotesOn, uint8(g.Scale[y]+(12*2)))
					}
				}
				g.SetPlayheadPosition()
				g.Play()
			}
		}
	}()
}

// SetPlayheadPosition ...
func (g *Graph) SetPlayheadPosition() {

	g.Playhead.Imd.Clear()
	g.Playhead.X = g.X + (float64(g.BeatIndex) * g.W / float64(g.XSteps))
	g.Playhead.Compose()
}

// Play ...
func (g *Graph) Play() {
	for _, n := range g.NotesOn {
		writer.NoteOn(g.MidiWriter, n, 100)
	}
}

// Stop ...
func (g *Graph) Stop() {
	for _, n := range g.NotesOn {
		writer.NoteOff(g.MidiWriter, n)
	}
	g.NotesOn = g.NotesOn[:0] // Keep allocated memory: To keep the underlying array, slice the slice to zero length
}
