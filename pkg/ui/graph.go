package ui

import (
	"fmt"
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

// Point ...
type Point struct {
	x, y uint32
}

// Graph ...
type Graph struct {
	Rect           pixel.Rect
	W, H           float64
	Matrix         [][]uint32
	UserBlocks     []Point
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
	Bpm            uint32
}

// NewGraph ...
func NewGraph(r pixel.Rect, ao midi.Out) *Graph {

	g := new(Graph)

	g.Rect = r
	g.W = g.Rect.W()
	g.H = g.Rect.H()

	g.Imd = imdraw.New(nil)

	// Initialize playhead
	g.Playhead = NewPlayhead(pixel.R(g.Rect.Min.X, g.Rect.Min.Y, g.Rect.Min.X, g.Rect.Max.Y))
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
	g.Bpm = 120
}

// Compose ...
func (g *Graph) Compose() {

	// Reset Matrix
	g.Matrix = make([][]uint32, int(g.XSteps))
	for i := range g.Matrix {
		g.Matrix[i] = make([]uint32, int(g.YSteps))
	}

	for _, point := range g.UserBlocks {
		if int(point.x) < len(g.Matrix) && int(point.y) < len(g.Matrix[0]) {
			g.Matrix[point.x][point.y] = 1
		}
	}

	xPos := uint32(0)
	for xPos < g.XSteps {
		val := simplexnoise.Fbm(float32(xPos+g.Offset), 0, g.Frequency, g.Lacunarity, g.Gain, int(g.Octaves))
		yPos := uint32(math.Round(helpers.ReRange(float64(val), -1, 1, 0, float64(g.YSteps-1))))
		g.Matrix[xPos][yPos] = 1
		xPos++
	}

	g.Imd.Clear()

	// Background
	g.Imd.Color = color.RGBA{0xee, 0xee, 0xee, 0xff}
	g.Imd.Push(
		pixel.V(g.Rect.Min.X, g.Rect.Min.Y),
		pixel.V(g.Rect.Max.X, g.Rect.Max.Y),
	)
	g.Imd.Rectangle(0)

	blockWidth := g.W / float64(g.XSteps)
	blockHeight := g.H / float64(g.YSteps)

	for x := range g.Matrix {
		for y := range g.Matrix[x] {

			// Draw active blocks
			if g.Matrix[x][y] == 1 {
				g.Imd.Color = color.RGBA{0x00, 0x00, 0x00, 0xff}
				g.Imd.Push(
					pixel.V(
						g.Rect.Min.X+(float64(x)*blockWidth),
						g.Rect.Min.Y+(float64(y)*blockHeight),
					),
					pixel.V(
						g.Rect.Min.X+(float64(x)*blockWidth)+blockWidth,
						g.Rect.Min.Y+(float64(y)*blockHeight)+blockHeight,
					),
				)
				g.Imd.Rectangle(0)
			}
		}
	}

	for x := 0; x <= len(g.Matrix); x++ {
		// Vertical Lines
		g.Imd.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
		g.Imd.Push(
			pixel.V(
				g.Rect.Min.X+(float64(x)*blockWidth),
				g.Rect.Min.Y,
			),
			pixel.V(
				g.Rect.Min.X+(float64(x)*blockWidth),
				g.Rect.Max.Y,
			),
		)
		g.Imd.Line(1)
	}

	for y := 0; y <= len(g.Matrix[0]); y++ {
		// Horizontal Lines
		g.Imd.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
		g.Imd.Push(
			pixel.V(
				g.Rect.Min.X,
				g.Rect.Min.Y+(float64(y)*blockHeight),
			),
			pixel.V(
				g.Rect.Max.X,
				g.Rect.Min.Y+(float64(y)*blockHeight),
			),
		)
		g.Imd.Line(1)
	}
}

// DrawTo ...
func (g *Graph) DrawTo(imd *imdraw.IMDraw) {
	g.Imd.Draw(imd)
	g.Playhead.DrawTo(imd)
}

// RespondToInput ...
func (g *Graph) RespondToInput(win *pixelgl.Window) {

	if win.Pressed(pixelgl.MouseButtonLeft) {
		pos := win.MousePosition()
		if helpers.PosInBounds(pos, g.Rect) {
			fmt.Println(pos)
			x := uint32((pos.X - g.Rect.Min.X) / (g.W / float64(g.XSteps)))
			y := uint32((pos.Y - g.Rect.Min.Y) / (g.H / float64(g.YSteps)))
			if g.Matrix[x][y] == 0 {
				// Add
				g.UserBlocks = append(g.UserBlocks, Point{x, y})
			}
			g.Compose()
		}
	}

	if win.Pressed(pixelgl.MouseButtonRight) {
		pos := win.MousePosition()
		if helpers.PosInBounds(pos, g.Rect) {
			fmt.Println(pos)
			x := uint32((pos.X - g.Rect.Min.X) / (g.W / float64(g.XSteps)))
			y := uint32((pos.Y - g.Rect.Min.Y) / (g.H / float64(g.YSteps)))
			for i, point := range g.UserBlocks {
				if point.x == x && point.y == y {
					// Delete
					g.UserBlocks = append(g.UserBlocks[:i], g.UserBlocks[i+1:]...)
				}
			}
			g.Compose()
		}
	}

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
	g.Playhead.Rect.Min.X = g.Rect.Min.X + (float64(g.BeatIndex) * g.W / float64(g.XSteps))
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
