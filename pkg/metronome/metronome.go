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
	CtrlChannel  chan signals.CtrlSignal
}

// New creates a new instance of Metronome
func New(bpm uint32) *Metronome {

	period := bpmToPeriod(bpm)

	m := &Metronome{
		Period: period,
		Ticker: time.NewTicker(period),
	}

	m.CtrlChannel = make(chan signals.CtrlSignal)
	m.ListenToCtrlChannel()

	return m
}

// SetBpm converts bpm to a time.Duration and updates the current period
func (m *Metronome) SetBpm(bpm uint32) {
	period := bpmToPeriod(bpm)
	m.SetPeriod(period)
}

// SetPeriod updates the current period
func (m *Metronome) SetPeriod(period time.Duration) {
	m.Period = period
	m.Ticker.Reset(period)
}

// AddBeatChannel ...
func (m *Metronome) AddBeatChannel(beatChannel chan signals.BeatSignal) {
	m.BeatChannels = append(m.BeatChannels, beatChannel)
}

// Start ...
func (m *Metronome) Start() {
	go func() {
		for {
			select {
			case <-m.Ticker.C:
				m.BeatSignal.Value = 1
				// Send the beat to all BeatChannels
				for index := range m.BeatChannels {
					m.BeatChannels[index] <- m.BeatSignal
				}
			}
		}
	}()
}

// ListenToCtrlChannel ...
func (m *Metronome) ListenToCtrlChannel() {
	go func() {
		for {
			select {
			case ctrlSignal := <-m.CtrlChannel:
				switch ctrlSignal.Label {
				case "reset":
					m.SetBpm(120)
				case "bpm":
					m.SetBpm(uint32(ctrlSignal.Value))
				}
			}
		}
	}()
}

func bpmToPeriod(bpm uint32) time.Duration {
	return time.Duration(60000/bpm) * time.Millisecond
}
