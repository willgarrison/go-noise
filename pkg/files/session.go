package files

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/willgarrison/go-noise/pkg/signals"
)

var lock sync.Mutex

type Session struct {
	InputCtrlChannel chan signals.Signal
	OutputChannels   []chan signals.Signal
	SessionData      SessionData
}

// SessionData ...
type SessionData struct {
	UserMatrix [][]uint32
	Frequency  float32
	Lacunarity float32
	Gain       float32
	Octaves    uint8
	XSteps     uint32
	YSteps     uint32
	Offset     uint32
	Bpm        uint32
	BeatLength uint32
}

func NewSession() *Session {

	s := new(Session)

	s.InputCtrlChannel = make(chan signals.Signal)
	s.ListenToInputCtrlChannel()

	s.InitSessionData()

	return s
}

func (s *Session) InitSessionData() {
	s.SessionData.Frequency = 0.3
	s.SessionData.Lacunarity = 0.9
	s.SessionData.Gain = 2.0
	s.SessionData.Octaves = 5
	s.SessionData.XSteps = 16
	s.SessionData.YSteps = 24
	s.SessionData.Offset = 0
	s.SessionData.Bpm = 180
	s.SessionData.BeatLength = 1

	// Initialize UserMatrix
	s.SessionData.UserMatrix = make([][]uint32, 64)
	for i := range s.SessionData.UserMatrix {
		s.SessionData.UserMatrix[i] = make([]uint32, 48)
	}
}

// Save saves a representation of v to the file at path.
func (s *Session) Save(path string) error {

	lock.Lock()
	defer lock.Unlock()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := json.Marshal(s.SessionData)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, bytes.NewReader(b))
	return err
}

// Load ...
func (s *Session) Load(path string) error {

	lock.Lock()
	defer lock.Unlock()

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(&s.SessionData)
}

// ListenToInputCtrlChannel ...
func (s *Session) ListenToInputCtrlChannel() {
	go func() {
		for {
			signal := <-s.InputCtrlChannel
			switch signal.Label {
			case "reset":
				s.InitSessionData()
				signal := signals.Signal{
					Label: "reset",
				}
				s.SendToOutputChannels(signal)
			case "save":
				fmt.Println("saving...")
				err := s.Save("test.json")
				if err != nil {
					log.Print(err)
				} else {
					fmt.Println("saved")
					signal := signals.Signal{
						Label: "saved",
					}
					s.SendToOutputChannels(signal)
				}
			case "load":
				fmt.Println("loading...")
				err := s.Load("test.json")
				if err != nil {
					log.Print(err)
				} else {
					fmt.Println("loading complete")
					signal := signals.Signal{
						Label: "loaded",
					}
					s.SendToOutputChannels(signal)
				}
			default:
			}
		}
	}()
}

// AddOutputChannel ...
func (s *Session) AddOutputChannel(outputChannel chan signals.Signal) {
	s.OutputChannels = append(s.OutputChannels, outputChannel)
}

// SendToOutputChannels ...
func (s *Session) SendToOutputChannels(signal signals.Signal) {
	// Send ctrl signal to all subscribers
	for index := range s.OutputChannels {
		s.OutputChannels[index] <- signal
	}
}
