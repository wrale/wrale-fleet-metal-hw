package gpio

import (
	"sync"

	"periph.io/x/conn/v3/gpio"
)

// pwmState tracks PWM pin state
type pwmState struct {
	pin       gpio.PinIO
	config    PWMConfig
	enabled   bool
	dutyCycle uint32
	mux       sync.Mutex
	done      chan struct{}
	wg        sync.WaitGroup
}
