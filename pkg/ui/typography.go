package ui

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/gomono"
)

// Typography ...
type Typography struct {
	txtBatch *pixel.Batch
	txt      *text.Text
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

	typ.txtBatch = pixel.NewBatch(&pixel.TrianglesData{}, txtAtlas.Picture())
	typ.txt = text.New(pixel.ZV, txtAtlas)

	return typ
}
