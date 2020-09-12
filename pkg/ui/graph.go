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

// Graph ...
type Graph struct {
	X, Y, W, H     float64
	Matrix         [][]uint32
	Imd            *imdraw.IMDraw
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
func NewGraph(r pixel.Rect) *Graph {

	g := new(Graph)

	g.X = r.Min.X
	g.Y = r.Min.Y
	g.W = r.Max.X
	g.H = r.Max.Y

	g.Imd = imdraw.New(nil)

	g.Reset()

	return g
}

// Reset ...
func (g *Graph) Reset() {
	g.Frequency = 0.3
	g.Lacunarity = 0.5
	g.Gain = 2.0
	g.Octaves = 5
	g.XSteps = 32
	g.YSteps = 48
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
				x1 := float64(x) * g.W / float64(g.XSteps)
				y1 := float64(y) * g.H / float64(g.YSteps)
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
}

// RespondToInput ...
func (g *Graph) RespondToInput(win *pixelgl.Window) {
	if g.SignalReceived {
		g.SignalReceived = false
		g.Compose()
	}
}

// Listen ...
func (g *Graph) Listen(graphChan chan signals.ControlValue) {
	for {
		select {
		case msg := <-graphChan:

			switch msg.Label {
			case "reset":
				g.Reset()
			case "frequency":
				g.Frequency = float32(msg.Value)
			case "lacunarity":
				g.Lacunarity = float32(msg.Value)
			case "gain":
				g.Gain = float32(msg.Value)
			case "octaves":
				g.Octaves = uint8(msg.Value)
			case "xSteps":
				g.XSteps = uint32(msg.Value)
			case "ySteps":
				g.YSteps = uint32(msg.Value)
			case "offset":
				g.Offset = uint32(msg.Value)
			}

			g.SignalReceived = true
		}
	}
}
