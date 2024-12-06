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

// updatePWM updates the PWM output for a pin
func (c *Controller) updatePWM(state *pwmState) error {
	if !state.enabled {
		return nil
	}

	if c.simulation {
		return nil // No actual pin updates in simulation mode
	}
	
	// Update duty cycle immediately
	if state.dutyCycle == 0 {
		return state.pin.Out(gpio.Low)
	}
	if state.dutyCycle == 100 {
		return state.pin.Out(gpio.High)
	}

	return nil // PWM loop will handle intermediate values
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
	if state.enabled {
		return c.updatePWM(state)
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
		
		// Set pin low after goroutine exits
		return state.pin.Out(gpio.Low)
	}

	return nil
}

// GetPWMState returns the current PWM configuration
func (c *Controller) GetPWMState(name string) (PWMConfig, error) {
	c.mux.RLock()
	state, exists := c.pwmPins[name]
	c.mux.RUnlock()

	if !exists {
		return PWMConfig{}, fmt.Errorf("PWM pin %s not found", name)
	}

	state.mux.Lock()
	defer state.mux.Unlock()

	return state.config, nil
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
			state.mux.Unlock()

			if dutyCycle == 0 {
				state.pin.Out(gpio.Low)
				timer.Reset(period)
				continue
			}
			if dutyCycle == 100 {
				state.pin.Out(gpio.High)
				timer.Reset(period)
				continue
			}

			onTime := period * time.Duration(dutyCycle) / 100
			offTime := period - onTime

			state.pin.Out(gpio.High)
			time.Sleep(onTime)
			state.pin.Out(gpio.Low)
			timer.Reset(offTime)
		}
	}
}