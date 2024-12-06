package gpio

import (
	"fmt"
	"sync"
	"time"

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

// ConfigurePWM sets up a pin for PWM operation
func (c *Controller) ConfigurePWM(name string, pin gpio.PinIO, cfg PWMConfig) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if !c.enabled {
		return fmt.Errorf("GPIO controller is disabled")
	}

	// Validate config
	if cfg.Frequency == 0 {
		cfg.Frequency = 25000 // Default to 25kHz
	}
	if cfg.DutyCycle > 100 {
		return fmt.Errorf("duty cycle must be 0-100")
	}

	if c.simulation {
		// In simulation mode, allow nil pin and just track state
		c.pwmPins[name] = &pwmState{
			pin:       pin,
			config:    cfg,
			enabled:   false,
			dutyCycle: cfg.DutyCycle,
			done:      make(chan struct{}),
		}
		// Store pull configuration in simulated state
		c.simPins[name] = &simPin{
			value: false,
			pull:  cfg.Pull,
		}
		return nil
	}

	if pin == nil {
		return fmt.Errorf("pin cannot be nil in non-simulation mode")
	}

	// Configure pin for output with pull setting
	if err := pin.In(cfg.Pull, gpio.NoEdge); err != nil {
		return fmt.Errorf("failed to configure pin pull: %w", err)
	}
	if err := pin.Out(gpio.Low); err != nil {
		return fmt.Errorf("failed to configure pin as output: %w", err)
	}

	// Create PWM state
	c.pwmPins[name] = &pwmState{
		pin:       pin,
		config:    cfg,
		enabled:   false,
		dutyCycle: cfg.DutyCycle,
		done:      make(chan struct{}),
	}

	return nil
}

// The rest of pwm.go remains unchanged
