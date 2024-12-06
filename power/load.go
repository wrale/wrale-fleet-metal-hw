package power

import (
	"context"
	"fmt"
	"time"
)

// LoadTestConfig configures power load testing parameters
type LoadTestConfig struct {
	// Target current draw in amps
	TargetCurrent float64
	
	// Test duration
	Duration time.Duration
	
	// Minimum required voltage during test
	MinVoltage float64
	
	// Maximum allowed voltage ripple
	MaxRipple float64
	
	// Callback for test progress
	OnProgress func(LoadTestProgress)
}

// LoadTestProgress represents current test status
type LoadTestProgress struct {
	CurrentDraw float64
	Voltage     float64
	Ripple      float64
	Duration    time.Duration
	Complete    bool
	Error       error
}

// RunLoadTest performs a power supply load test
// Tests if power supply can maintain stable voltage at specified current
func (m *Manager) RunLoadTest(ctx context.Context, cfg LoadTestConfig) error {
	if cfg.TargetCurrent <= 0 {
		return fmt.Errorf("invalid target current: must be > 0")
	}
	if cfg.Duration <= 0 {
		return fmt.Errorf("invalid duration: must be > 0")
	}
	if cfg.MinVoltage <= 0 {
		return fmt.Errorf("invalid minimum voltage: must be > 0")
	}

	// Get initial state
	initialState := m.GetState()
	
	// Setup test load GPIO pins
	loadPins := []string{
		"load_bank_1",
		"load_bank_2", 
		"load_bank_3",
	}
	
	// Enable load banks gradually
	for _, pin := range loadPins {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := m.gpio.SetPinState(pin, true); err != nil {
				return fmt.Errorf("failed to enable load bank: %w", err)
			}
			
			// Wait for current to stabilize
			time.Sleep(100 * time.Millisecond)
			
			// Check current draw
			state := m.GetState()
			current := state.PowerConsumption / state.Voltage
			
			if current >= cfg.TargetCurrent {
				break
			}
		}
	}
	
	startTime := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var maxRipple float64
	var minVoltage = initialState.Voltage
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
			
		case <-ticker.C:
			state := m.GetState()
			current := state.PowerConsumption / state.Voltage
			
			// Update min voltage seen
			if state.Voltage < minVoltage {
				minVoltage = state.Voltage
			}
			
			// Get voltage ripple from stability monitor if available
			if state.StabilityMetrics != nil && 
			   state.StabilityMetrics.VoltageRipple > maxRipple {
				maxRipple = state.StabilityMetrics.VoltageRipple
			}
			
			// Report progress
			if cfg.OnProgress != nil {
				cfg.OnProgress(LoadTestProgress{
					CurrentDraw: current,
					Voltage:     state.Voltage,
					Ripple:      maxRipple,
					Duration:    time.Since(startTime),
					Complete:    false,
				})
			}
			
			// Check test completion
			if time.Since(startTime) >= cfg.Duration {
				// Disable load banks
				for _, pin := range loadPins {
					m.gpio.SetPinState(pin, false)
				}
				
				// Validate test results
				if minVoltage < cfg.MinVoltage {
					return fmt.Errorf("voltage dropped below minimum: %v < %v", 
						minVoltage, cfg.MinVoltage)
				}
				if maxRipple > cfg.MaxRipple {
					return fmt.Errorf("voltage ripple exceeded maximum: %v > %v",
						maxRipple, cfg.MaxRipple)
				}
				if current < cfg.TargetCurrent {
					return fmt.Errorf("failed to reach target current: %v < %v",
						current, cfg.TargetCurrent)
				}
				
				return nil
			}
			
			// Verify voltage hasn't dropped too low
			if state.Voltage < cfg.MinVoltage {
				return fmt.Errorf("voltage dropped below minimum: %v", state.Voltage)
			}
		}
	}
}