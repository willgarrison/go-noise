package ui

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func NewWindow(title string, w, h float64) *pixelgl.Window {

	config := pixelgl.WindowConfig{
		Title:     title,
		Bounds:    pixel.R(0, 0, w, h),
		Resizable: false,
		VSync:     true,
	}

	win, err := pixelgl.NewWindow(config)
	if err != nil {
		panic(err)
	}

	return win
}
