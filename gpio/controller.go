package gpio

import (
	"fmt"
	"sync"
	"time"

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

// EnablePWM starts PWM output on a pin
func (c *Controller) EnablePWM(name string) error {
	c.mux.RLock()
	state, exists := c.pwmPins[name]
	c.mux.RUnlock()

	if !exists {
		return fmt.Errorf("PWM pin %s not found", name)
	}

	state.mux.Lock()
	defer state.mux.Unlock()

	if state.enabled {
		return nil // Already enabled
	}

	state.enabled = true
	state.done = make(chan struct{})

	if !c.simulation {
		state.wg.Add(1)
		go c.pwmLoop(state)
	}

	return nil
}

// DisablePWM stops PWM output on a pin
func (c *Controller) DisablePWM(name string) error {
	c.mux.RLock()
	state, exists := c.pwmPins[name]
	c.mux.RUnlock()

	if !exists {
		return fmt.Errorf("PWM pin %s not found", name)
	}

	state.mux.Lock()
	if !state.enabled {
		state.mux.Unlock()
		return nil // Already disabled
	}

	state.enabled = false
	close(state.done)
	state.mux.Unlock()

	if !c.simulation {
		// Wait for PWM loop to exit
		state.wg.Wait()

		if state.pin != nil {
			// Set pin low after goroutine exits
			if err := state.pin.Out(gpio.Low); err != nil {
				return fmt.Errorf("failed to set pin low: %w", err)
			}
		}
	}

	return nil
}

// SetPWMDutyCycle updates the PWM duty cycle (0-100)
func (c *Controller) SetPWMDutyCycle(name string, dutyCycle uint32) error {
	c.mux.RLock()
	state, exists := c.pwmPins[name]
	c.mux.RUnlock()

	if !exists {
		return fmt.Errorf("PWM pin %s not found", name)
	}

	state.mux.Lock()
	defer state.mux.Unlock()

	if dutyCycle > 100 {
		return fmt.Errorf("duty cycle must be 0-100")
	}

	state.dutyCycle = dutyCycle

	// Update PWM if enabled
	if state.enabled && !c.simulation {
		if state.pin != nil {
			if dutyCycle == 0 {
				if err := state.pin.Out(gpio.Low); err != nil {
					return fmt.Errorf("failed to set pin low: %w", err)
				}
			}
			if dutyCycle == 100 {
				if err := state.pin.Out(gpio.High); err != nil {
					return fmt.Errorf("failed to set pin high: %w", err)
				}
			}
		}
	}

	return nil
}

// pwmLoop handles the PWM signal generation
func (c *Controller) pwmLoop(state *pwmState) {
	defer state.wg.Done()
	period := time.Duration(1000000000/state.config.Frequency) * time.Nanosecond

	timer := time.NewTimer(period)
	defer timer.Stop()

	for {
		select {
		case <-state.done:
			return
		case <-timer.C:
			state.mux.Lock()
			if !state.enabled {
				state.mux.Unlock()
				return
			}

			dutyCycle := state.dutyCycle
			pin := state.pin
			state.mux.Unlock()

			if pin == nil {
				continue
			}

			if dutyCycle == 0 {
				if err := pin.Out(gpio.Low); err != nil {
					// Log error but continue - PWM is best-effort
					continue
				}
				timer.Reset(period)
				continue
			}
			if dutyCycle == 100 {
				if err := pin.Out(gpio.High); err != nil {
					// Log error but continue - PWM is best-effort
					continue
				}
				timer.Reset(period)
				continue
			}

			onTime := period * time.Duration(dutyCycle) / 100
			offTime := period - onTime

			if err := pin.Out(gpio.High); err != nil {
				// Log error but continue - PWM is best-effort
				continue
			}
			time.Sleep(onTime)
			if err := pin.Out(gpio.Low); err != nil {
				// Log error but continue - PWM is best-effort
				continue
			}
			timer.Reset(offTime)
		}
	}
}

// Close releases all GPIO resources
func (c *Controller) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	var lastErr error

	if !c.simulation {
		// Disable all PWM outputs
		for name := range c.pwmPins {
			if err := c.DisablePWM(name); err != nil {
				lastErr = err
			}
		}

		// Set all pins to safe state
		for _, pin := range c.pins {
			if pin != nil {
				if err := pin.Out(gpio.Low); err != nil {
					lastErr = err
				}
			}
		}
	}

	c.enabled = false
	return lastErr
}
