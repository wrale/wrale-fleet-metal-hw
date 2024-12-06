package gpio

import "periph.io/x/conn/v3/gpio"

// Pull specifies the pull up/down state for GPIO pins
type Pull = gpio.Pull

const (
	// PullNone specifies no pull up/down
	PullNone = gpio.Float
)

// PWMConfig holds PWM pin configuration
type PWMConfig struct {
	// PWM frequency in Hz
	Frequency uint32

	// Initial duty cycle (0-100)
	DutyCycle uint32

	// Pull up/down configuration
	Pull gpio.Pull
}

// Options configures GPIO controller behavior
type Options struct {
	// SimulationMode bypasses hardware initialization
	SimulationMode bool
}

// Option is a function that configures Options
type Option func(*Options)

// WithSimulation enables simulation mode
func WithSimulation() Option {
	return func(opts *Options) {
		opts.SimulationMode = true
	}
}
