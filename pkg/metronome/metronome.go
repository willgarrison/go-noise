package metronome

import (
	"fmt"
	"time"

	"github.com/willgarrison/go-noise/pkg/session"
	"github.com/willgarrison/go-noise/pkg/signals"
)

// Metronome ...
type Metronome struct {
	Period              time.Duration
	Ticker              *time.Ticker
	OutputChannels      []chan signals.Signal
	InputCtrlChannel    chan signals.Signal
	InputSessionChannel chan signals.Signal
	SessionData         *session.SessionData
}

// New creates a new instance of Metronome
func New(sessionData *session.SessionData) *Metronome {

	period := bpmToPeriod(sessionData.Bpm)

	m := &Metronome{
		Period:      period,
		Ticker:      time.NewTicker(period),
		SessionData: sessionData,
	}

	m.InputCtrlChannel = make(chan signals.Signal)
	m.ListenToInputCtrlChannel()

	m.InputSessionChannel = make(chan signals.Signal)
	m.ListenToInputSessionChannel()

	return m
}

// SetBpm converts bpm to a time.Duration and updates the current period
func (m *Metronome) SetBpm(bpm uint32) {
	m.SessionData.Bpm = bpm
	period := bpmToPeriod(bpm)
	m.SetPeriod(period)
}

// SetPeriod updates the current period
func (m *Metronome) SetPeriod(period time.Duration) {
	m.Period = period
	m.Ticker.Reset(period)
}

// Start ...
func (m *Metronome) Start() {
	go func() {
		for {
			<-m.Ticker.C
			signal := signals.Signal{
				Value: 1,
			}
			m.SendToOutputChannels(signal)
		}
	}()
}

// ListenToInputCtrlChannel ...
func (m *Metronome) ListenToInputCtrlChannel() {
	go func() {
		for {
			ctrlSignal := <-m.InputCtrlChannel
			switch ctrlSignal.Label {
			case "reset":
				m.SetBpm(180)
			case "bpm":
				m.SetBpm(uint32(ctrlSignal.Value))
			default:
			}
		}
	}()
}

// ListenToInputSessionChannel ...
func (m *Metronome) ListenToInputSessionChannel() {
	go func() {
		for {
			signal := <-m.InputSessionChannel
			switch signal.Label {
			case "saved":
				fmt.Println("metronome: session data saved")
			case "loaded":
				fmt.Println("metronome: update from session data")
			default:
			}
		}
	}()
}

// AddOutputChannel ...
func (m *Metronome) AddOutputChannel(outputChannel chan signals.Signal) {
	m.OutputChannels = append(m.OutputChannels, outputChannel)
}

// SendToOutputChannels ...
func (m *Metronome) SendToOutputChannels(signal signals.Signal) {
	// Send ctrl signal to all subscribers
	for index := range m.OutputChannels {
		m.OutputChannels[index] <- signal
	}
}

func bpmToPeriod(bpm uint32) time.Duration {
	return time.Duration(60000/bpm) * time.Millisecond
}
