package gpio

import (
	"fmt"
	"sync"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
)

// simPin tracks simulated pin state
type simPin struct {
	value bool
	pull  gpio.Pull
}

// Controller manages GPIO pins and their states
type Controller struct {
	mux        sync.RWMutex
	pins       map[string]gpio.PinIO
	interrupts map[string]*interruptState
	pwmPins    map[string]*pwmState
	enabled    bool
	simulation bool
	
	// Simulated state
	simPins map[string]*simPin
}

// New creates a new GPIO controller
func New(opts ...Option) (*Controller, error) {
	// Process options
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	// Initialize host for GPIO access if not in simulation mode
	if !options.SimulationMode {
		if _, err := host.Init(); err != nil {
			return nil, fmt.Errorf("failed to initialize GPIO host: %w", err)
		}
	}

	return &Controller{
		pins:       make(map[string]gpio.PinIO),
		interrupts: make(map[string]*interruptState),
		pwmPins:    make(map[string]*pwmState),
		enabled:    true,
		simulation: options.SimulationMode,
		simPins:    make(map[string]*simPin),
	}, nil
}

// ConfigurePin sets up a GPIO pin for use with optional pull-up/down
func (c *Controller) ConfigurePin(name string, pin gpio.PinIO, pull gpio.Pull) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if !c.enabled {
		return fmt.Errorf("GPIO controller is disabled")
	}

	if c.simulation {
		c.pins[name] = pin // Allow nil pin in simulation mode
		c.simPins[name] = &simPin{
			value: false,
			pull:  pull,
		}
		// Even in simulation mode, configure the pin if one was provided
		if pin != nil {
			if err := pin.In(pull, gpio.NoEdge); err != nil {
				return fmt.Errorf("failed to configure pin: %w", err)
			}
		}
		return nil
	}

	if pin == nil {
		return fmt.Errorf("pin cannot be nil in non-simulation mode")
	}

	// Configure pin for input with pull setting
	if err := pin.In(pull, gpio.NoEdge); err != nil {
		return fmt.Errorf("failed to configure pin: %w", err)
	}

	c.pins[name] = pin
	return nil
}

// SetPinState sets the state of a GPIO pin
func (c *Controller) SetPinState(name string, high bool) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.simulation {
		simPin, exists := c.simPins[name]
		if !exists {
			return fmt.Errorf("pin %s not found", name)
		}
		simPin.value = high
		// Also set physical pin if one exists
		if pin := c.pins[name]; pin != nil {
			if high {
				return pin.Out(gpio.High)
			}
			return pin.Out(gpio.Low)
		}
		return nil
	}

	pin, exists := c.pins[name]
	if !exists {
		return fmt.Errorf("pin %s not found", name)
	}

	if pin == nil {
		return fmt.Errorf("pin %s is nil", name)
	}

	if high {
		return pin.Out(gpio.High)
	}
	return pin.Out(gpio.Low)
}

// GetPinState reads the current state of a GPIO pin
func (c *Controller) GetPinState(name string) (bool, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	if c.simulation {
		simPin, exists := c.simPins[name]
		if !exists {
			return false, fmt.Errorf("pin %s not found", name)
		}
		// Also read physical pin if one exists
		if pin := c.pins[name]; pin != nil {
			return pin.Read() == gpio.High, nil
		}
		return simPin.value, nil
	}

	pin, exists := c.pins[name]
	if !exists {
		return false, fmt.Errorf("pin %s not found", name)
	}

	if pin == nil {
		return false, fmt.Errorf("pin %s is nil", name)
	}

	return pin.Read() == gpio.High, nil
}

// GetPinPull reads the current pull-up/down configuration of a GPIO pin
func (c *Controller) GetPinPull(name string) (gpio.Pull, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	if c.simulation {
		simPin, exists := c.simPins[name]
		if !exists {
			return gpio.Float, fmt.Errorf("pin %s not found", name)
		}
		return simPin.pull, nil
	}

	pin, exists := c.pins[name]
	if !exists {
		return gpio.Float, fmt.Errorf("pin %s not found", name)
	}

	if pin == nil {
		return gpio.Float, fmt.Errorf("pin %s is nil", name)
	}

	return pin.Pull(), nil
}

// Close releases all GPIO resources
func (c *Controller) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if !c.simulation {
		// Disable all PWM outputs
		for name := range c.pwmPins {
			c.DisablePWM(name)
		}

		// Set all pins to safe state
		for _, pin := range c.pins {
			if pin != nil {
				pin.Out(gpio.Low)
			}
		}
	}

	c.enabled = false
	return nil
}