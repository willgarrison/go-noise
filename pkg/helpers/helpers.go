package helpers

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
)

// ReRange maps a value from one range to another
func ReRange(value, aMin, aMax, bMin, bMax float64) float64 {

	if value < aMin {
		value = aMin
	}

	if value > aMax {
		value = aMax
	}

	newValue := (((value - aMin) * (bMax - bMin)) / (aMax - aMin)) + bMin
	return newValue
}

// Constrain caps the range (low and high) of a given float64 (n)
func Constrain(n, low, high float64) float64 {
	switch {
	case n < low:
		return low
	case n > high:
		return high
	default:
		return n
	}
}

// PosInBounds ...
func PosInBounds(pos pixel.Vec, bounds []pixel.Vec) bool {
	if pos.X > bounds[0].X &&
		pos.Y > bounds[0].Y &&
		pos.X < bounds[1].X &&
		pos.Y < bounds[1].Y {
		return true
	}
	return false
}

// DrawTextToBatch ...
func DrawTextToBatch(s string, vec pixel.Vec, txtBatch *pixel.Batch, txt *text.Text) {
	txt.Clear()
	txt.Color = color.RGBA{0x00, 0x00, 0x00, 0xff}
	txt.Dot = vec
	txt.WriteString(s)
	txt.Draw(txtBatch, pixel.IM)
}
