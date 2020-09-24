package ui

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/gomono"
)

// Typography ...
type Typography struct {
	TxtBatch *pixel.Batch
	Txt      *text.Text
}

// NewTypography ...
func NewTypography() *Typography {

	typ := new(Typography)

	// Font
	ttf, err := truetype.Parse(gomono.TTF)
	if err != nil {
		panic(err)
	}
	fontFace := truetype.NewFace(ttf, &truetype.Options{Size: 10})
	txtAtlas := text.NewAtlas(fontFace, text.ASCII)

	typ.TxtBatch = pixel.NewBatch(&pixel.TrianglesData{}, txtAtlas.Picture())
	typ.Txt = text.New(pixel.ZV, txtAtlas)

	return typ
}

// DrawTextToBatch ...
func (typ *Typography) DrawTextToBatch(s string, vec pixel.Vec, clr color.Color, txtBatch *pixel.Batch, txt *text.Text) {
	txt.Clear()
	txt.Color = clr
	txt.Dot = vec
	txt.WriteString(s)
	txt.Draw(txtBatch, pixel.IM)
}
