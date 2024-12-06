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