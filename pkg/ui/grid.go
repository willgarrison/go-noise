package ui

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strconv"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/willgarrison/go-noise/pkg/generators"
	"github.com/willgarrison/go-noise/pkg/helpers"
	"github.com/willgarrison/go-noise/pkg/session"
	"github.com/willgarrison/go-noise/pkg/signals"
	"github.com/willgarrison/go-noise/pkg/simplexnoise"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/writer"
)

type Note struct {
	index       uint8
	release     uint8
	beatsPlayed uint8
	isPlaying   bool
}

type Grid struct {
	Rect                pixel.Rect
	W, H                float64
	Imd                 *imdraw.IMDraw
	Matrix              [][]uint32
	InputBeatChannel    chan signals.Signal
	InputCtrlChannel    chan signals.Signal
	InputSessionChannel chan signals.Signal
	BeatIndex           uint8
	Notes               []Note
	NotesToStrike       []uint8
	Scale               []uint8
	NoteNames           []string
	MidiWriter          *writer.Writer
	MidiOutput          midi.Out
	Playhead            *Playhead
	Typ                 *Typography
	IsPlaying           bool
	SignalReceived      bool
	SessionData         *session.SessionData
}

func NewGrid(r pixel.Rect, ao midi.Out, sessionData *session.SessionData) *Grid {

	g := new(Grid)

	g.Rect = r
	g.W = g.Rect.W()
	g.H = g.Rect.H()

	g.SessionData = sessionData

	g.Imd = imdraw.New(nil)

	// Initialize notes
	g.Notes = make([]Note, 128)
	for i := range g.Notes {
		g.Notes[i].index = uint8(i)
		g.Notes[i].release = uint8(g.SessionData.Release)
	}

	// Initialize playhead
	g.Playhead = NewPlayhead(pixel.R(g.Rect.Min.X, g.Rect.Min.Y, g.Rect.Min.X, g.Rect.Max.Y))
	g.Playhead.Compose()

	g.SetScale(0)

	g.MidiWriter = writer.New(ao)
	g.MidiWriter.SetChannel(1)

	g.InputCtrlChannel = make(chan signals.Signal)
	g.ListenToInputCtrlChannel()

	g.InputBeatChannel = make(chan signals.Signal)
	g.ListenToInputBeatChannel()

	g.InputSessionChannel = make(chan signals.Signal)
	g.ListenToInputSessionChannel()

	g.Typ = NewTypography()

	return g
}

func (g *Grid) Compose() {

	// Reset Matrix
	g.Matrix = make([][]uint32, int(g.SessionData.XSteps))
	for i := range g.Matrix {
		g.Matrix[i] = make([]uint32, int(g.SessionData.YSteps))
	}

	xPos := uint32(0)
	for xPos < g.SessionData.XSteps {
		val := simplexnoise.Fbm(float32(xPos+g.SessionData.Offset), 0, float32(g.SessionData.Frequency), float32(g.SessionData.Lacunarity), float32(g.SessionData.Gain), int(g.SessionData.Octaves))
		yPos := uint32(math.Round(helpers.ReRange(float64(val), -1, 1, 0, float64(g.SessionData.YSteps-1))))
		g.Matrix[xPos][yPos] = 1
		xPos++
	}

	// Set beatlength
	for i := range g.Notes {
		g.Notes[i].release = g.SessionData.Release
	}

	// Set active blocks
	for x := range g.Matrix {
		for y := range g.Matrix[x] {
			if g.SessionData.UserMatrix[x][y] != 0 {
				g.Matrix[x][y] = g.SessionData.UserMatrix[x][y]
			}
		}
	}

	g.SessionData.UserPattern, _ = generators.NewEuclid(g.SessionData.N, g.SessionData.K, g.SessionData.R, g.SessionData.G)

	// Clear
	g.Imd.Clear()
	g.Typ.TxtBatch.Clear()

	backgroundColumnColor := color.RGBA{0xdd, 0xdd, 0xdd, 0xff}
	activeColumnColor := color.RGBA{0xee, 0xee, 0xee, 0xff}

	// Background
	g.Imd.Color = backgroundColumnColor
	g.Imd.Push(
		pixel.V(g.Rect.Min.X, g.Rect.Min.Y),
		pixel.V(g.Rect.Max.X, g.Rect.Max.Y),
	)
	g.Imd.Rectangle(0)

	blockWidth := g.W / float64(g.SessionData.XSteps)
	blockHeight := g.H / float64(g.SessionData.YSteps)

	rhythmLength := len(g.SessionData.UserPattern.Rhythm)

	// Draw active columns and blocks
	for x := range g.Matrix {

		// Draw active columns
		if g.SessionData.UserPattern.Rhythm[x%rhythmLength] == 1 {
			g.Imd.Color = activeColumnColor
			g.Imd.Push(
				pixel.V(
					g.Rect.Min.X+(float64(x)*blockWidth),
					g.Rect.Min.Y,
				),
				pixel.V(
					g.Rect.Min.X+(float64(x)*blockWidth)+blockWidth,
					g.Rect.Max.Y,
				),
			)
			g.Imd.Rectangle(0)
		}

		// Draw active blocks
		for y := range g.Matrix[x] {
			if g.Matrix[x][y] > 0 {

				// System block
				blockColor := color.RGBA{0x10, 0x10, 0x10, 0xff}

				// User block
				if g.Matrix[x][y] == 2 { // On
					blockColor = color.RGBA{0x36, 0xaf, 0xcf, 0xff}
				} else if g.Matrix[x][y] == 3 { // Off
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
	for x := 0; x <= len(g.Matrix); x++ {
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
	for y := 0; y <= len(g.Matrix[0]); y++ {
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
	for x := 0; x < len(g.Matrix); x++ {
		str := strconv.Itoa(x + 1)
		strX := g.Rect.Min.X + (float64(x) * blockWidth) + (blockWidth / 2) - (g.Typ.Txt.BoundsOf(str).W() / 2)
		strY := g.Rect.Min.Y - (g.Typ.Txt.BoundsOf(str).H() + 10)
		g.Typ.DrawTextToBatch(str, pixel.V(strX, strY), color.RGBA{0x00, 0x00, 0x00, 0xff}, g.Typ.TxtBatch, g.Typ.Txt)
	}

	// Text: Notes
	for y := 0; y < len(g.Matrix[0]); y++ {
		midiNote := strconv.Itoa(int(g.SessionData.Low) + y)
		noteName := g.NoteNames[(g.Scale[y]+g.SessionData.Low)%12]
		noteNameX := g.Rect.Min.X - (g.Typ.Txt.BoundsOf(midiNote+" "+noteName).W() + 20)
		noteNameY := g.Rect.Min.Y + (float64(y) * blockHeight) + (blockHeight / 2) - (g.Typ.Txt.BoundsOf(noteName).H() / 3)
		g.Typ.DrawTextToBatch(midiNote+" "+noteName, pixel.V(noteNameX, noteNameY), color.RGBA{0x00, 0x00, 0x00, 0xff}, g.Typ.TxtBatch, g.Typ.Txt)
	}
}

func (g *Grid) DrawTo(imd *imdraw.IMDraw) {
	g.Imd.Draw(imd)
	g.Playhead.DrawTo(imd)
}

func (g *Grid) RespondToInput(win *pixelgl.Window) {

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

func (g *Grid) SetScale(scaleIndex int) {

	// C   Db  D   Eb  E   F   F#  G   Ab  A   Bb   B
	// 0   1   2   3   4   5   6   7   8   9   10   11
	scales := [][]uint8{
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, // 12 Tone
		{0, 2, 4, 5, 7, 9, 11},                 // Major
		{0, 2, 3, 5, 7, 8, 10},                 // Natural minor	C, D, Eb, F, G, Ab, Bb
		{0, 2, 3, 5, 7, 8, 11},                 // Harmonic minor	C, D, Eb, F, G, Ab, B
		{0, 2, 3, 5, 7, 9, 11},                 // Melodic minor	C, D, Eb, F, G, A, B
		{0, 2, 4, 7, 9},                        // Pentatonic 		C, D, E, G, A, C
	}

	scaleIndex = scaleIndex % len(scales)

	var octave uint8
	var midiNote uint8
	octave = 0
	midiNote = 0

	g.Scale = []uint8{}
	for midiNote < 127 {
		for n := range scales[scaleIndex] {
			midiNote = (12 * octave) + scales[scaleIndex][n]
			g.Scale = append(g.Scale, midiNote)
		}
		octave++
	}

	g.NoteNames = []string{"C", "Db", "D", "Eb", "E", "F", "F#", "G", "Ab", "A", "Bb", "B"}
}

func (g *Grid) SetPlayheadPosition() {
	g.Playhead.Imd.Clear()
	g.Playhead.Rect.Min.X = g.Rect.Min.X + (float64(g.BeatIndex) * g.W / float64(g.SessionData.XSteps))
	g.Playhead.Compose()
}

func (g *Grid) TurnNotesOn() {
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

func (g *Grid) TurnNotesOff() {
	for i, note := range g.Notes {
		if note.isPlaying {
			g.Notes[i].beatsPlayed++
			if g.Notes[i].beatsPlayed >= note.release {
				writer.NoteOff(g.MidiWriter, note.index)
				g.Notes[i].beatsPlayed = 0
				g.Notes[i].isPlaying = false
			}
		}
	}
}

func (g *Grid) TurnAllNotesOff() {
	for i, note := range g.Notes {
		writer.NoteOff(g.MidiWriter, note.index)
		g.Notes[i].beatsPlayed = 0
		g.Notes[i].isPlaying = false
	}
}

func (g *Grid) Play() {
	g.IsPlaying = true
}

func (g *Grid) Stop() {
	g.IsPlaying = false
	g.BeatIndex = 0
	g.TurnAllNotesOff()
	g.SetPlayheadPosition()
}

func (g *Grid) Toggle() {
	if g.IsPlaying {
		g.Stop()
	} else {
		g.Play()
	}
}

func (g *Grid) ListenToInputCtrlChannel() {
	go func() {
		for {
			signal := <-g.InputCtrlChannel
			switch signal.Label {
			case "12 tone":
				g.SetScale(0)
			case "major":
				g.SetScale(1)
			case "natural":
				g.SetScale(2)
			case "harmonic":
				g.SetScale(3)
			case "melodic":
				g.SetScale(4)
			case "pentatonic":
				g.SetScale(5)
			case "play":
				g.Play()
			case "stop":
				g.Stop()
			case "toggle":
				g.Toggle()
			case "freq":
				g.SessionData.Frequency = signal.Value
			case "space":
				g.SessionData.Lacunarity = signal.Value
			case "gain":
				g.SessionData.Gain = signal.Value
			case "octs":
				g.SessionData.Octaves = uint8(signal.Value)
			case "x":
				g.SessionData.XSteps = uint32(signal.Value)
			case "y":
				g.SessionData.YSteps = uint32(signal.Value)
			case "pos":
				g.SessionData.Offset = uint32(signal.Value)
			case "low":
				g.SessionData.Low = uint8(signal.Value)
			case "rel":
				g.SessionData.Release = uint8(signal.Value)
			case "n":
				g.SessionData.N = uint8(signal.Value)
			case "k":
				g.SessionData.K = uint8(signal.Value)
			case "r":
				g.SessionData.R = uint8(signal.Value)
			case "g":
				g.SessionData.G = signal.Value
			default:
			}
			g.SignalReceived = true
		}
	}()
}

func (g *Grid) ListenToInputBeatChannel() {
	go func() {
		for {
			beatSignal := <-g.InputBeatChannel
			if g.IsPlaying {
				if g.SessionData.UserPattern.Rhythm[g.BeatIndex%uint8(len(g.SessionData.UserPattern.Rhythm))] == 1 {
					for y, val := range g.Matrix[g.BeatIndex%uint8(len(g.Matrix))] {
						if val == 1 || val == 2 {
							note := helpers.ConstrainUInt8(g.SessionData.Low+g.Scale[y], 0, 127)
							g.NotesToStrike = append(g.NotesToStrike, note)
						}
					}
				}
				g.TurnNotesOff()
				g.TurnNotesOn()
				g.SetPlayheadPosition()
				g.BeatIndex = (g.BeatIndex + uint8(beatSignal.Value)) % uint8(len(g.Matrix))
			}
		}
	}()
}

func (g *Grid) ListenToInputSessionChannel() {
	go func() {
		for {
			signal := <-g.InputSessionChannel
			switch signal.Label {
			case "reset":
				fmt.Println("grid: session data reset")
			case "saved":
				fmt.Println("grid: session data saved")
			case "loaded":
				fmt.Println("grid: update from session data")
			default:
			}
			g.SignalReceived = true
		}
	}()
}
