package ui

import (
	"encoding/json"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"strconv"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/willgarrison/go-noise/pkg/helpers"
	"github.com/willgarrison/go-noise/pkg/signals"
	"github.com/willgarrison/go-noise/pkg/simplexnoise"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/writer"
)

// Note ...
type Note struct {
	index       uint8
	beatLength  uint8
	beatsPlayed uint8
	isPlaying   bool
}

// Point ...
type Point struct {
	x, y uint32
}

// Graph ...
type Graph struct {
	Rect           pixel.Rect
	W, H           float64
	Imd            *imdraw.IMDraw
	CtrlChannel    chan signals.CtrlSignal
	BeatChannel    chan signals.BeatSignal
	BeatIndex      uint8
	Notes          []Note
	NotesToStrike  []uint8
	Scale          []uint8
	NoteNames      []string
	MidiWriter     *writer.Writer
	MidiOutput     midi.Out
	Playhead       *Playhead
	Typ            *Typography
	IsPlaying      bool
	SignalReceived bool
	SessionData    SessionData
}

// SessionData ...
type SessionData struct {
	Matrix     [][]uint32
	UserMatrix [][]uint32
	Frequency  float32
	Lacunarity float32
	Gain       float32
	Octaves    uint8
	XSteps     uint32
	YSteps     uint32
	Offset     uint32
	Bpm        uint32
	BeatLength uint32
}

// NewGraph ...
func NewGraph(r pixel.Rect, ao midi.Out) *Graph {

	g := new(Graph)

	g.Rect = r
	g.W = g.Rect.W()
	g.H = g.Rect.H()

	g.Imd = imdraw.New(nil)

	// Initialize notes
	g.Notes = make([]Note, 128)
	for i := range g.Notes {
		g.Notes[i].index = uint8(i)
		g.Notes[i].beatLength = uint8(g.SessionData.BeatLength)
	}

	// Initialize playhead
	g.Playhead = NewPlayhead(pixel.R(g.Rect.Min.X, g.Rect.Min.Y, g.Rect.Min.X, g.Rect.Max.Y))
	g.Playhead.Compose()

	g.Reset()

	g.SetScale(0)

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
	g.SessionData.Frequency = 0.3
	g.SessionData.Lacunarity = 0.9
	g.SessionData.Gain = 2.0
	g.SessionData.Octaves = 5
	g.SessionData.XSteps = 16
	g.SessionData.YSteps = 24
	g.SessionData.Offset = 0
	g.SessionData.Bpm = 180
	g.SessionData.BeatLength = 1

	// Initialize UserMatrix
	g.SessionData.UserMatrix = make([][]uint32, 64)
	for i := range g.SessionData.UserMatrix {
		g.SessionData.UserMatrix[i] = make([]uint32, 48)
	}
}

// Compose ...
func (g *Graph) Compose() {

	// Reset Matrix
	g.SessionData.Matrix = make([][]uint32, int(g.SessionData.XSteps))
	for i := range g.SessionData.Matrix {
		g.SessionData.Matrix[i] = make([]uint32, int(g.SessionData.YSteps))
	}

	xPos := uint32(0)
	for xPos < g.SessionData.XSteps {
		val := simplexnoise.Fbm(float32(xPos+g.SessionData.Offset), 0, g.SessionData.Frequency, g.SessionData.Lacunarity, g.SessionData.Gain, int(g.SessionData.Octaves))
		yPos := uint32(math.Round(helpers.ReRange(float64(val), -1, 1, 0, float64(g.SessionData.YSteps-1))))
		g.SessionData.Matrix[xPos][yPos] = 1
		xPos++
	}

	// Set beatlength
	for i := range g.Notes {
		g.Notes[i].beatLength = uint8(g.SessionData.BeatLength)
	}

	// Draw active blocks
	for x := range g.SessionData.Matrix {
		for y := range g.SessionData.Matrix[x] {
			if g.SessionData.UserMatrix[x][y] != 0 {
				g.SessionData.Matrix[x][y] = g.SessionData.UserMatrix[x][y]
			}
		}
	}

	// Clear
	g.Imd.Clear()
	g.Typ.TxtBatch.Clear()

	// Background
	g.Imd.Color = color.RGBA{0xee, 0xee, 0xee, 0xff}
	g.Imd.Push(
		pixel.V(g.Rect.Min.X, g.Rect.Min.Y),
		pixel.V(g.Rect.Max.X, g.Rect.Max.Y),
	)
	g.Imd.Rectangle(0)

	blockWidth := g.W / float64(g.SessionData.XSteps)
	blockHeight := g.H / float64(g.SessionData.YSteps)

	// Draw active blocks
	for x := range g.SessionData.Matrix {
		for y := range g.SessionData.Matrix[x] {
			if g.SessionData.Matrix[x][y] > 0 {

				// System block
				blockColor := color.RGBA{0x00, 0x00, 0x00, 0xff}

				// User block
				if g.SessionData.Matrix[x][y] == 2 { // On
					blockColor = color.RGBA{0x36, 0xaf, 0xcf, 0xff}
				} else if g.SessionData.Matrix[x][y] == 3 { // Off
					blockColor = color.RGBA{0x90, 0x90, 0x90, 0xff}
				}

				g.Imd.Color = blockColor
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

	// Vertical Lines
	for x := 0; x <= len(g.SessionData.Matrix); x++ {
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

	// Horizontal Lines
	for y := 0; y <= len(g.SessionData.Matrix[0]); y++ {
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

	// Text: Beats
	for x := 0; x < len(g.SessionData.Matrix); x++ {
		str := strconv.Itoa(x + 1)
		strX := g.Rect.Min.X + (float64(x) * blockWidth) + (blockWidth / 2) - (g.Typ.Txt.BoundsOf(str).W() / 2)
		strY := g.Rect.Min.Y - (g.Typ.Txt.BoundsOf(str).H() + 10)
		g.Typ.DrawTextToBatch(str, pixel.V(strX, strY), color.RGBA{0x00, 0x00, 0x00, 0xff}, g.Typ.TxtBatch, g.Typ.Txt)
	}

	// Text: Notes
	for y := 0; y < len(g.SessionData.Matrix[0]); y++ {
		str := g.NoteNames[g.Scale[y]%12]
		strX := g.Rect.Min.X - (g.Typ.Txt.BoundsOf(str).W() + 10)
		strY := g.Rect.Min.Y + (float64(y) * blockHeight) + (blockHeight / 2) - (g.Typ.Txt.BoundsOf(str).H() / 3)
		g.Typ.DrawTextToBatch(str, pixel.V(strX, strY), color.RGBA{0x00, 0x00, 0x00, 0xff}, g.Typ.TxtBatch, g.Typ.Txt)
	}
}

// DrawTo ...
func (g *Graph) DrawTo(imd *imdraw.IMDraw) {
	g.Imd.Draw(imd)
	g.Playhead.DrawTo(imd)
}

// RespondToInput ...
func (g *Graph) RespondToInput(win *pixelgl.Window) {

	if win.JustPressed(pixelgl.MouseButtonLeft) {
		pos := win.MousePosition()
		if helpers.PosInBounds(pos, g.Rect) {
			x := uint32((pos.X - g.Rect.Min.X) / (g.W / float64(g.SessionData.XSteps)))
			y := uint32((pos.Y - g.Rect.Min.Y) / (g.H / float64(g.SessionData.YSteps)))
			if g.SessionData.UserMatrix[x][y] == 2 {
				g.SessionData.UserMatrix[x][y] = 0
			} else {
				g.SessionData.UserMatrix[x][y] = 2
			}
			g.Compose()
		}
	}

	if win.JustPressed(pixelgl.MouseButtonRight) {
		pos := win.MousePosition()
		if helpers.PosInBounds(pos, g.Rect) {
			x := uint32((pos.X - g.Rect.Min.X) / (g.W / float64(g.SessionData.XSteps)))
			y := uint32((pos.Y - g.Rect.Min.Y) / (g.H / float64(g.SessionData.YSteps)))
			if g.SessionData.UserMatrix[x][y] == 3 {
				g.SessionData.UserMatrix[x][y] = 0
			} else {
				g.SessionData.UserMatrix[x][y] = 3
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
				case "toggle":
					g.Toggle()
				case "save":
					g.Save()
				case "load":
					g.Load()
				case "freq":
					g.SessionData.Frequency = float32(ctrlSignal.Value)
				case "space":
					g.SessionData.Lacunarity = float32(ctrlSignal.Value)
				case "gain":
					g.SessionData.Gain = float32(ctrlSignal.Value)
				case "octs":
					g.SessionData.Octaves = uint8(ctrlSignal.Value)
				case "x":
					g.SessionData.XSteps = uint32(ctrlSignal.Value)
				case "y":
					g.SessionData.YSteps = uint32(ctrlSignal.Value)
				case "pos":
					g.SessionData.Offset = uint32(ctrlSignal.Value)
				case "rel":
					g.SessionData.BeatLength = uint32(ctrlSignal.Value)
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
					for y, val := range g.SessionData.Matrix[g.BeatIndex%uint8(len(g.SessionData.Matrix))] {
						if val == 1 || val == 2 {
							g.NotesToStrike = append(g.NotesToStrike, uint8(g.Scale[y]+(12*2)))
						}
					}
					g.TurnNotesOff()
					g.TurnNotesOn()
					g.SetPlayheadPosition()
					g.BeatIndex = (g.BeatIndex + beatSignal.Value) % uint8(len(g.SessionData.Matrix))
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
	g.Playhead.Rect.Min.X = g.Rect.Min.X + (float64(g.BeatIndex) * g.W / float64(g.SessionData.XSteps))
	g.Playhead.Compose()
}

// TurnNotesOn ...
func (g *Graph) TurnNotesOn() {
	for _, note := range g.NotesToStrike {
		// If already playing, turn off
		if g.Notes[note].isPlaying {
			writer.NoteOff(g.MidiWriter, note)
		}
		// Turn on
		writer.NoteOn(g.MidiWriter, note, uint8(rand.Intn(50)+51))
		g.Notes[note].beatsPlayed = 0
		g.Notes[note].isPlaying = true
		// Clean up, but keep allocated memory
		// To keep the underlying array, slice the slice to zero length
		g.NotesToStrike = g.NotesToStrike[:0]
	}
}

// TurnNotesOff ...
func (g *Graph) TurnNotesOff() {
	for i, note := range g.Notes {
		if note.isPlaying {
			g.Notes[i].beatsPlayed++
			if g.Notes[i].beatsPlayed >= note.beatLength {
				writer.NoteOff(g.MidiWriter, note.index)
				g.Notes[i].beatsPlayed = 0
				g.Notes[i].isPlaying = false
			}
		}
	}
}

// TurnAllNotesOff ...
func (g *Graph) TurnAllNotesOff() {
	for i, note := range g.Notes {
		writer.NoteOff(g.MidiWriter, note.index)
		g.Notes[i].beatsPlayed = 0
		g.Notes[i].isPlaying = false
	}
}

// Play ...
func (g *Graph) Play() {
	g.IsPlaying = true
}

// Stop ...
func (g *Graph) Stop() {
	g.IsPlaying = false
	g.BeatIndex = 0
	g.TurnAllNotesOff()
	g.SetPlayheadPosition()
}

// Toggle ...
func (g *Graph) Toggle() {
	if g.IsPlaying {
		g.Stop()
	} else {
		g.Play()
	}
}

// Save ...
func (g *Graph) Save() {

	jsonData, err := json.Marshal(g.SessionData)
	if err != nil {
		log.Println(err)
	}

	err = ioutil.WriteFile("test.json", jsonData, 0644)
	if err != nil {
		log.Println(err)
	}
}

// Load ...
func (g *Graph) Load() {

	jsonData, err := ioutil.ReadFile("test.json")
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(jsonData, &g.SessionData)
	if err != nil {
		log.Println(err)
	}
}
