package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gen2brain/dlgs"
	"github.com/willgarrison/go-noise/pkg/generators"
	"github.com/willgarrison/go-noise/pkg/helpers"
	"github.com/willgarrison/go-noise/pkg/signals"
)

var lock sync.Mutex

type Session struct {
	InputCtrlChannel chan signals.Signal
	OutputChannels   []chan signals.Signal
	SessionData      SessionData
}

type SessionData struct {
	UserMatrix       [][]uint32
	UserPattern      *generators.Pattern
	KeyboardNumInput string
	Frequency        float64
	Lacunarity       float64
	Gain             float64
	Octaves          uint8
	XSteps           uint32
	YSteps           uint32
	Offset           uint32
	Bpm              uint32
	Low              uint8
	Release          uint8
	N, K, R          uint8 // Pattern Variables
	G                float64
}

func NewSession() *Session {

	s := new(Session)

	s.InputCtrlChannel = make(chan signals.Signal)
	s.ListenToInputCtrlChannel()

	s.InitializeSessionData()

	return s
}

func (s *Session) InitializeSessionData() {

	s.SessionData.UserMatrix = make([][]uint32, 64)
	for i := range s.SessionData.UserMatrix {
		s.SessionData.UserMatrix[i] = make([]uint32, 48)
	}

	// set a random seed
	rand.Seed(time.Now().UnixNano())

	// Set default parameters to random values
	s.SessionData.Frequency = 0.3
	s.SessionData.Lacunarity = 0.9
	s.SessionData.Gain = helpers.RandFloatInRange(1.5, 3.0)
	s.SessionData.Octaves = uint8(helpers.RandIntInRange(3, 6))

	// valid x steps
	xSteps := []uint32{4, 8, 16, 24, 32}
	s.SessionData.XSteps = xSteps[rand.Intn(len(xSteps))]

	// valid y steps
	ySteps := []uint32{8, 12, 24, 32}
	s.SessionData.YSteps = ySteps[rand.Intn(len(ySteps))]

	s.SessionData.Offset = uint32(helpers.RandIntInRange(1, 999))
	s.SessionData.Bpm = 180
	s.SessionData.Low = 36
	s.SessionData.Release = 1
	s.SessionData.N = 16
	s.SessionData.K = 16
	s.SessionData.R = 0
	s.SessionData.G = 0

	s.SessionData.UserPattern, _ = generators.NewEuclid(s.SessionData.N, s.SessionData.K, s.SessionData.R, s.SessionData.G)
}

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

func (s *Session) ListenToInputCtrlChannel() {
	go func() {
		for {
			signal := <-s.InputCtrlChannel
			switch signal.Label {
			case "reset":

				s.InitializeSessionData()

				signal := signals.Signal{
					Label: "reset",
				}

				s.SendToOutputChannels(signal)

			case "save":

				defaultFilename := "session-" + strconv.Itoa(int(time.Now().Unix()))
				enteredFileName, _, err := dlgs.Entry("Save Session", "Save as:", defaultFilename)
				if err != nil {
					log.Println("dlgs.Entry:", err)
				}

				directory, _, err := dlgs.File("Select a save location:", "", true)
				if err != nil {
					log.Println("dlgs.File:", err)
				}

				err = s.Save(directory + "/" + enteredFileName + ".json")
				if err != nil {
					log.Println("s.Save:", err)
				} else {
					fmt.Println("saved")
					signal := signals.Signal{
						Label: "saved",
					}
					s.SendToOutputChannels(signal)
				}

			case "load":

				selectedFile, _, err := dlgs.File("Select file:", "", false)
				if err != nil {
					log.Println("dlgs.File:", err)
				}

				err = s.Load(selectedFile)
				if err != nil {
					log.Println("s.Load:", err)
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

func (s *Session) AddOutputChannel(outputChannel chan signals.Signal) {
	s.OutputChannels = append(s.OutputChannels, outputChannel)
}

func (s *Session) SendToOutputChannels(signal signals.Signal) {
	// Send ctrl signal to all subscribers
	for index := range s.OutputChannels {
		s.OutputChannels[index] <- signal
	}
}
