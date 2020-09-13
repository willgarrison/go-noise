package metronome

import (
	"time"
)

// Metronome ...
type Metronome struct {
	Period       time.Duration
	Ticker       *time.Ticker
	Beat         uint8
	BeatChannels []chan uint8
}

// New creates a new instance of Metronome
func New(bpm uint16) *Metronome {

	period := bpmToPeriod(bpm)

	mt := &Metronome{
		Period: period,
		Ticker: time.NewTicker(period),
	}

	return mt
}

// AddBeatChannel ...
func (mt *Metronome) AddBeatChannel(bc chan uint8) {
	mt.BeatChannels = append(mt.BeatChannels, bc)
}

// Start ...
func (mt *Metronome) Start() {
	go func() {
		for {
			select {
			case <-mt.Ticker.C:
				// Send the beat to all BeatChannels
				for index := range mt.BeatChannels {
					mt.BeatChannels[index] <- mt.Beat
				}
				// Increment the beat
				mt.Beat++
			}
		}
	}()
}

func bpmToPeriod(bpm uint16) time.Duration {
	return time.Duration(60000/bpm) * time.Millisecond
}
