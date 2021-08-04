package generators

import "github.com/willgarrison/go-noise/pkg/helpers"

type Pattern struct {
	Rhythm []uint8
}

func NewEuclid(n, k, rotation uint8, groove float64) (*Pattern, error) {

	p := new(Pattern)

	// flip n and k if n is greater than k
	if n > k {
		n, k = k, n
	}

	p.createEuclidPattern(n, k)

	if groove != 0 {
		p.setGroove(n, k, groove)
	}

	if rotation != 0 {
		p.setRotate(int(rotation))
	}

	return p, nil
}

// createPattern creates a new rhythmic pattern using Bresenham’s line algorithm
func (p *Pattern) createEuclidPattern(n, k uint8) {

	p.Rhythm = []uint8{}

	previous := -1

	ratio := float64(n) / float64(k)

	var i uint8
	for i < k {
		x := int(ratio * float64(i))
		if x != previous {
			p.Rhythm = append(p.Rhythm, 1)
		} else {
			p.Rhythm = append(p.Rhythm, 0)
		}
		previous = x
		i++
	}
}

func (p *Pattern) setRotate(rotation int) {

	np := []uint8{}

	offset := 0

	// If offset is negative
	if rotation < 0 {
		// subtract from length,
		// the positive of offset,
		// constrained to length via mod
		offset = len(p.Rhythm) - ((rotation * -1) % len(p.Rhythm))
	} else {
		offset = rotation % len(p.Rhythm)
	}

	np = append(np, p.Rhythm[offset:]...)
	np = append(np, p.Rhythm[:offset]...)

	p.Rhythm = np
}

func (p *Pattern) setGroove(n, k uint8, groove float64) {

	tmpRhythm := make([]uint8, len(p.Rhythm))
	copy(tmpRhythm, p.Rhythm)

	gn := (uint8(helpers.ReRange(groove, 0, 100, 0, float64(k))) + n)
	if gn > k {
		gn = k
	}

	groovePattern := new(Pattern)
	groovePattern.createEuclidPattern(gn, k)

	midPointIndex := int(k / 2)

	i := 1 // skip the first beat
	for i < len(p.Rhythm) {

		if p.Rhythm[i] == 1 && i != midPointIndex {

			tmpRhythm[i] = 0

			groovePatternIndex := i
			distance := 1
			direction := 1 // positive is forward, negative backwards

			found := false

			for !found {

				groovePatternIndex = (groovePatternIndex + (distance * direction)) % len(groovePattern.Rhythm)

				if groovePatternIndex < 0 {
					tmpRhythm[i] = 1
					break
				}

				if groovePattern.Rhythm[groovePatternIndex] == 1 && tmpRhythm[groovePatternIndex] != 1 {
					tmpRhythm[groovePatternIndex] = 1
					found = true
				}

				distance++
				direction *= -1
			}
		}

		i++
	}

	p.Rhythm = tmpRhythm
}
