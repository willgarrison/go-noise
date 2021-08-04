package helpers

import (
	"github.com/faiface/pixel"
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

func ConstrainFloat64(n, low, high float64) float64 {
	switch {
	case n < low:
		return low
	case n > high:
		return high
	default:
		return n
	}
}

func ConstrainUInt8(n, low, high uint8) uint8 {
	switch {
	case n < low:
		return low
	case n > high:
		return high
	default:
		return n
	}
}

func PosInBounds(pos pixel.Vec, rect pixel.Rect) bool {
	return rect.Contains(pos)
}
