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
	NoteNames      []string
	MidiWriter     *writer.Writer
	MidiOutput     midi.Out
	Playhead       *Playhead
	Typ            *Typography
	IsPlaying      bool
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

	g.SetScale(1)

	g.MidiWriter = writer.New(ao)
	g.MidiWriter.SetChannel(1)

	g.CtrlChannel = make(chan signals.CtrlSignal)
	g.ListenToCtrlChannel()

	g.BeatChannel = make(chan signals.BeatSignal)
	g.ListenToBeatChannel()

	g.Typ = NewTypography()

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

	g.Typ.TxtBatch.Clear()

	for y := 0; y < len(g.Matrix[0]); y++ {

		// Notes
		str := g.NoteNames[g.Scale[y]%12]
		strX := g.Rect.Min.X - (g.Typ.Txt.BoundsOf(str).W() + 10)
		strY := g.Rect.Min.Y + (float64(y) * blockHeight) + (blockHeight / 2) - (g.Typ.Txt.BoundsOf(str).H() / 3)
		g.Typ.DrawTextToBatch(str, pixel.V(strX, strY), g.Typ.TxtBatch, g.Typ.Txt)
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
				case "major":
					g.SetScale(0)
				case "natural":
					g.SetScale(1)
				case "harmonic":
					g.SetScale(2)
				case "melodic":
					g.SetScale(3)
				case "pentatonic":
					g.SetScale(4)
				case "play":
					g.Play()
				case "stop":
					g.Stop()
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
				if g.IsPlaying {
					g.BeatIndex = beatSignal.Value % uint8(len(g.Matrix))
					g.StopNotes()
					for y, val := range g.Matrix[g.BeatIndex] {
						if val == 1 {
							g.NotesOn = append(g.NotesOn, uint8(g.Scale[y]+(12*2)))
						}
					}
					g.SetPlayheadPosition()
					g.PlayNotes()
				}
			}
		}
	}()
}

// SetScale ...
func (g *Graph) SetScale(scaleIndex int) {

	g.NoteNames = []string{"C", "Db", "D", "Eb", "E", "F", "F#", "G", "Ab", "A", "Bb", "B"}

	// C   Db  D   Eb  E   F   F#  G   Ab  A   Bb   B
	// 0   1   2   3   4   5   6   7   8   9   10   11
	scales := [][]uint8{
		{0, 2, 4, 5, 7, 9, 11}, // Major
		{0, 2, 3, 5, 7, 8, 10}, // Natural minor	C, D, Eb, F, G, Ab, Bb
		{0, 2, 3, 5, 7, 8, 11}, // Harmonic minor	C, D, Eb, F, G, Ab, B
		{0, 2, 3, 5, 7, 9, 11}, // Melodic minor	C, D, Eb, F, G, A, B
		{0, 2, 4, 7, 9},        // Pentatonic 		C, D, E, G, A, C
	}

	realScaleIndex := scaleIndex % len(scales)

	fmt.Println("realScaleIndex:", realScaleIndex)

	g.Scale = []uint8{}
	for i := 0; i < 12; i++ {
		for n := range scales[realScaleIndex] {
			g.Scale = append(g.Scale, scales[realScaleIndex][n]+uint8(12*i))
		}
	}
}

// SetPlayheadPosition ...
func (g *Graph) SetPlayheadPosition() {
	g.Playhead.Imd.Clear()
	g.Playhead.Rect.Min.X = g.Rect.Min.X + (float64(g.BeatIndex) * g.W / float64(g.XSteps))
	g.Playhead.Compose()
}

// PlayNotes ...
func (g *Graph) PlayNotes() {
	for _, n := range g.NotesOn {
		writer.NoteOn(g.MidiWriter, n, 100)
	}
}

// StopNotes ...
func (g *Graph) StopNotes() {
	for _, n := range g.NotesOn {
		writer.NoteOff(g.MidiWriter, n)
	}
	g.NotesOn = g.NotesOn[:0] // Keep allocated memory: To keep the underlying array, slice the slice to zero length
}

// Play ...
func (g *Graph) Play() {
	g.IsPlaying = true
}

// Stop ...
func (g *Graph) Stop() {
	g.IsPlaying = false
}
