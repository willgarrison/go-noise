package midi

import (
	"gitlab.com/gomidi/midi"
	driver "gitlab.com/gomidi/rtmididrv"
)

// Midi ...
type Midi struct {
	Driver *driver.Driver
	Output midi.Out
}

// New ...
func New() (*Midi, error) {

	m := &Midi{
		Output: nil,
	}

	// Create new driver
	drv, err := driver.New()
	if err != nil {
		return nil, err
	}

	// Set driver so it can be closed
	// from outside this package
	m.Driver = drv

	// Get outputs
	// outs, err := drv.Outs()
	// if err != nil {
	// 	return nil, err
	// }

	// Create and set output to new virtual out
	m.Output, err = m.Driver.OpenVirtualOut("NoiseVirtualOut")
	if err != nil {
		return nil, err
	}

	// Set output to default IAC Driver
	// m.Output = outs[0]

	// Open output for writing
	err = m.Output.Open()
	if err != nil {
		return nil, err
	}

	return m, nil
}
