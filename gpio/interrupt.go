package gpio

import (
	"context"
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
)

// Edge represents interrupt trigger edges
type Edge string

const (
	Rising  Edge = "RISING"
	Falling Edge = "FALLING"
	Both    Edge = "BOTH"

	// Default debounce time for hardware interrupts
	defaultDebounceTime = 50 * time.Millisecond
)

// InterruptHandler is called when an interrupt occurs
type InterruptHandler func(pin string, state bool)

// InterruptConfig configures interrupt behavior
type InterruptConfig struct {
	Edge         Edge
	DebounceTime time.Duration
	Handler      InterruptHandler
}

// interruptState tracks interrupt configuration for a pin
type interruptState struct {
	config      InterruptConfig
	lastTrigger time.Time
	enabled     bool
}

// EnableInterrupt enables interrupt detection on a pin
func (c *Controller) EnableInterrupt(name string, cfg InterruptConfig) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	pin, exists := c.pins[name]
	if !exists {
		return fmt.Errorf("pin %s not found", name)
	}

	// Ensure pin supports interrupts
	input := pin.In(gpio.PullUp, gpio.BothEdges)
	if input != nil {
		return fmt.Errorf("failed to configure pin for interrupts: %v", input)
	}

	// Initialize interrupt tracking
	if c.interrupts == nil {
		c.interrupts = make(map[string]*interruptState)
	}

	// Set debounce time if not specified
	if cfg.DebounceTime == 0 {
		cfg.DebounceTime = defaultDebounceTime
	}

	// Store interrupt configuration
	c.interrupts[name] = &interruptState{
		config:  cfg,
		enabled: true,
	}

	return nil
}

// DisableInterrupt disables interrupt detection on a pin
func (c *Controller) DisableInterrupt(name string) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	state, exists := c.interrupts[name]
	if !exists {
		return fmt.Errorf("no interrupt configured for pin %s", name)
	}

	state.enabled = false
	return nil
}

// handleInterrupt processes a pin interrupt
func (c *Controller) handleInterrupt(name string, state bool) {
	c.mux.RLock()
	interrupt, exists := c.interrupts[name]
	if !exists || !interrupt.enabled {
		c.mux.RUnlock()
		return
	}

	// Check debounce
	now := time.Now()
	if now.Sub(interrupt.lastTrigger) < interrupt.config.DebounceTime {
		c.mux.RUnlock()
		return
	}

	// Update last trigger time
	interrupt.lastTrigger = now

	// Get handler
	handler := interrupt.config.Handler
	c.mux.RUnlock()

	// Call handler if configured
	if handler != nil {
		handler(name, state)
	}
}

// monitorPin monitors a single pin for interrupts
func (c *Controller) monitorPin(ctx context.Context, name string, pin gpio.PinIO) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Check pin state
			state := pin.Read() == gpio.High
			c.handleInterrupt(name, state)
			time.Sleep(time.Millisecond) // Prevent tight loop
		}
	}
}

// Monitor starts monitoring pin changes in the background
func (c *Controller) Monitor(ctx context.Context) error {
	// Start monitoring each pin with interrupts enabled
	for name, pin := range c.pins {
		if state, hasInterrupt := c.interrupts[name]; hasInterrupt && state.enabled {
			go c.monitorPin(ctx, name, pin)
		}
	}

	<-ctx.Done()
	return nil
}