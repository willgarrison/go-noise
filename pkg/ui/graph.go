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
)

var (
	frequency  float32 = 0.03
	lacunarity float32 = 0.5
	gain       float32 = 2.0
	octaves    float32 = 5.0
)

// Graph ...
type Graph struct {
	W              float64
	H              float64
	XSteps         uint32
	YSteps         uint32
	Matrix         [][]uint32
	Imd            *imdraw.IMDraw
	signalReceived bool
}

// NewGraph ...
func NewGraph(w, h float64) *Graph {

	g := new(Graph)

	g.W = w
	g.H = h

	g.XSteps = 300
	g.YSteps = 300

	g.Imd = imdraw.New(nil)

	return g
}

// Compose ...
func (g *Graph) Compose() {

	// Reset Matrix
	g.Matrix = make([][]uint32, g.XSteps)
	for i := range g.Matrix {
		g.Matrix[i] = make([]uint32, g.YSteps)
	}

	xPos := uint32(0)
	for xPos < g.XSteps {
		val := simplexnoise.Fbm(float32(xPos), 0, frequency, lacunarity, gain, int(octaves))
		yPos := uint32(math.Round(helpers.ReRange(float64(val), -1, 1, 0, float64(g.YSteps-1))))
		g.Matrix[xPos][yPos] = 1
		xPos++
	}

	g.Imd.Clear()

	// Background
	g.Imd.Color = color.RGBA{0xd0, 0xd0, 0xd0, 0xff}
	g.Imd.Push(
		pixel.V(0, 0),
		pixel.V(g.W, g.H),
	)
	g.Imd.Rectangle(0)

	// Draw active blocks
	g.Imd.Color = color.RGBA{0x00, 0x00, 0x00, 0xff}
	margin := 0.0
	blockWidth := g.W/float64(g.XSteps) - margin
	blockHeight := g.H/float64(g.YSteps) - margin
	for x := range g.Matrix {
		for y := range g.Matrix[x] {
			if g.Matrix[x][y] == 1 {
				g.Imd.Color = color.RGBA{0x00, 0x00, 0x00, 0xff}
				g.Imd.Push(
					pixel.V((float64(x)*g.W/float64(g.XSteps)+margin), (float64(y)*g.H/float64(g.YSteps)+margin)),
					pixel.V((float64(x)*g.W/float64(g.XSteps)+blockWidth), (float64(y)*g.H/float64(g.YSteps)+blockHeight)),
				)
				g.Imd.Rectangle(0)
			}
		}
	}
}

// Draw ...
func (g *Graph) Draw(win *pixelgl.Window) {
	g.Imd.Draw(win)
}

// RespondToInput ...
func (g *Graph) RespondToInput(win *pixelgl.Window) {
	if g.signalReceived {
		// fmt.Println("frequency:", frequency, "lacunarity:", lacunarity, "gain:", gain, "octaves:", octaves)
		g.signalReceived = false
		g.Compose()
	}
}

// Listen ...
func (g *Graph) Listen(graphChan chan signals.ControlValue) {
	for {
		select {
		case msg := <-graphChan:

			switch msg.Label {
			case "frequency":
				frequency = float32(msg.Value)
			case "lacunarity":
				lacunarity = float32(msg.Value)
			case "gain":
				gain = float32(msg.Value)
			case "octaves":
				octaves = float32(msg.Value)
			}

			g.signalReceived = true
		}
	}
}
