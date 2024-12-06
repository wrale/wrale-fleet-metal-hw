package gpio

import (
	"fmt"
	"sync"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
)

// Controller manages GPIO pins and their states
type Controller struct {
	mux        sync.RWMutex
	pins       map[string]gpio.PinIO
	interrupts map[string]*interruptState
	pwmPins    map[string]*pwmState
	enabled    bool
}

// New creates a new GPIO controller
func New() (*Controller, error) {
	// Initialize host for GPIO access
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize GPIO host: %w", err)
	}

	return &Controller{
		pins:       make(map[string]gpio.PinIO),
		interrupts: make(map[string]*interruptState),
		pwmPins:    make(map[string]*pwmState),
		enabled:    true,
	}, nil
}

// ConfigurePin sets up a GPIO pin for use with optional pull-up/down
func (c *Controller) ConfigurePin(name string, pin gpio.PinIO, pull gpio.Pull) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if !c.enabled {
		return fmt.Errorf("GPIO controller is disabled")
	}

	if err := pin.In(pull, gpio.NoEdge); err != nil {
		return fmt.Errorf("failed to configure pin: %w", err)
	}

	c.pins[name] = pin
	return nil
}

// SetPinState sets the state of a GPIO pin
func (c *Controller) SetPinState(name string, high bool) error {
	c.mux.RLock()
	defer c.mux.RUnlock()

	pin, exists := c.pins[name]
	if !exists {
		return fmt.Errorf("pin %s not found", name)
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

	pin, exists := c.pins[name]
	if !exists {
		return false, fmt.Errorf("pin %s not found", name)
	}

	return pin.Read() == gpio.High, nil
}

// Close releases all GPIO resources
func (c *Controller) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	// Disable all PWM outputs
	for name := range c.pwmPins {
		c.DisablePWM(name)
	}

	// Set all pins to safe state
	for _, pin := range c.pins {
		pin.Out(gpio.Low)
	}

	c.enabled = false
	return nil
}