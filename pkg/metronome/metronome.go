package metronome

import (
	"time"

	"github.com/willgarrison/go-noise/pkg/signals"
)

// Metronome ...
type Metronome struct {
	Period       time.Duration
	Ticker       *time.Ticker
	BeatSignal   signals.BeatSignal
	BeatChannels []chan signals.BeatSignal
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
func (mt *Metronome) AddBeatChannel(beatChannel chan signals.BeatSignal) {
	mt.BeatChannels = append(mt.BeatChannels, beatChannel)
}

// Start ...
func (mt *Metronome) Start() {
	go func() {
		for {
			select {
			case <-mt.Ticker.C:
				// Send the beat to all BeatChannels
				for index := range mt.BeatChannels {
					mt.BeatChannels[index] <- mt.BeatSignal
				}
				// Increment the beat
				mt.BeatSignal.Value++
			}
		}
	}()
}

func bpmToPeriod(bpm uint16) time.Duration {
	return time.Duration(60000/bpm) * time.Millisecond
}
