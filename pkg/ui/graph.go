package ui

import (
	"fmt"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/willgarrison/go-noise/pkg/helpers"
	"github.com/willgarrison/go-noise/pkg/simplexnoise"
	"golang.org/x/image/colornames"
)

var (
	frequency  float32 = 0.03
	lacunarity float32 = 0.5
	gain       float32 = 2.0
	octaves    int     = 5
)

// Graph ...
type Graph struct {
	W      float64
	H      float64
	XSteps uint32
	YSteps uint32
	Matrix [][]uint32
	Imd    *imdraw.IMDraw
}

// NewGraph ...
func NewGraph(w, h float64) *Graph {

	g := new(Graph)

	g.W = w
	g.H = h

	g.XSteps = 800
	g.YSteps = 600

	g.Imd = imdraw.New(nil)

	return g
}

// Generate ...
func (g *Graph) Generate() {

	// Reset Matrix
	g.Matrix = make([][]uint32, g.XSteps)
	for i := range g.Matrix {
		g.Matrix[i] = make([]uint32, g.YSteps)
	}

	xPos := uint32(0)
	for xPos < g.XSteps {
		val := simplexnoise.Fbm(float32(xPos), 0, frequency, lacunarity, gain, octaves)
		yPos := uint32(math.Round(helpers.ReRange(float64(val), -1, 1, 0, float64(g.YSteps-1))))
		g.Matrix[xPos][yPos] = 1
		xPos++
	}

	g.Imd.Clear()

	// Background
	g.Imd.Color = colornames.Lightgray
	g.Imd.Push(
		pixel.V(0, 0),
		pixel.V(g.W, g.H),
	)
	g.Imd.Rectangle(0)

	// Draw active blocks
	g.Imd.Color = colornames.Black
	margin := 0.0
	blockWidth := g.W/float64(g.XSteps) - margin
	blockHeight := g.H/float64(g.YSteps) - margin
	for x := range g.Matrix {
		for y := range g.Matrix[x] {
			if g.Matrix[x][y] == 1 {
				g.Imd.Color = colornames.Black
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

	redraw := false

	if win.Pressed(pixelgl.KeyUp) {
		switch {
		case win.Pressed(pixelgl.KeyF):
			frequency += 0.001
		case win.Pressed(pixelgl.KeyL):
			lacunarity += 0.01
		case win.Pressed(pixelgl.KeyG):
			gain += 0.1
		case win.Pressed(pixelgl.KeyO):
			octaves++
		}

		redraw = true
	}

	if win.Pressed(pixelgl.KeyDown) {
		switch {
		case win.Pressed(pixelgl.KeyF):
			frequency -= 0.001
		case win.Pressed(pixelgl.KeyL):
			lacunarity -= 0.01
		case win.Pressed(pixelgl.KeyG):
			gain -= 0.1
		case win.Pressed(pixelgl.KeyO):
			if octaves > 0 {
				octaves--
			}
		}

		redraw = true
	}

	if redraw {
		fmt.Println("frequency:", frequency, "lacunarity:", lacunarity, "gain:", gain, "octaves:", octaves)
		g.Generate()
	}
}
