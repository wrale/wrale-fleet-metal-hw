package thermal

import (
	"fmt"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

// Fan speed ranges and PWM settings
const (
	// Fan speed percentages as uint32 to avoid conversion
	fanSpeedLow    uint32 = 25
	fanSpeedMedium uint32 = 50
	fanSpeedHigh   uint32 = 100

	// PWM configuration
	fanPWMFrequency = 25000 // 25kHz standard for PC fans
)

// updateCooling adjusts cooling based on temperatures
func (m *Monitor) updateCooling() {
	m.mux.Lock()
	defer m.mux.Unlock()

	// Determine maximum temperature
	maxTemp := m.state.CPUTemp
	if m.state.GPUTemp > maxTemp {
		maxTemp = m.state.GPUTemp
	}

	// Calculate fan speed based on temperature ranges
	var dutyCycle uint32
	switch {
	case maxTemp >= cpuTempCritical:
		dutyCycle = fanSpeedHigh
		m.setThrottlingLocked(true)
	case maxTemp >= cpuTempWarning:
		// Linear interpolation between medium and high speed
		tempRange := cpuTempCritical - cpuTempWarning
		tempAboveWarning := maxTemp - cpuTempWarning
		speedRange := float64(fanSpeedHigh - fanSpeedMedium)
		interpolated := float64(fanSpeedMedium) + speedRange*(tempAboveWarning/tempRange)
		if interpolated < float64(fanSpeedMedium) {
			dutyCycle = fanSpeedMedium
		} else if interpolated > float64(fanSpeedHigh) {
			dutyCycle = fanSpeedHigh
		} else {
			dutyCycle = fanSpeedMedium + uint32(speedRange*(tempAboveWarning/tempRange))
		}
		m.setThrottlingLocked(false)
	case maxTemp >= (cpuTempWarning / 2):
		// Linear interpolation between low and medium speed
		tempRange := cpuTempWarning - (cpuTempWarning / 2)
		tempAboveMin := maxTemp - (cpuTempWarning / 2)
		speedRange := float64(fanSpeedMedium - fanSpeedLow)
		interpolated := float64(fanSpeedLow) + speedRange*(tempAboveMin/tempRange)
		if interpolated < float64(fanSpeedLow) {
			dutyCycle = fanSpeedLow
		} else if interpolated > float64(fanSpeedMedium) {
			dutyCycle = fanSpeedMedium
		} else {
			dutyCycle = fanSpeedLow + uint32(speedRange*(tempAboveMin/tempRange))
		}
		m.setThrottlingLocked(false)
	default:
		dutyCycle = fanSpeedLow
		m.setThrottlingLocked(false)
	}

	// Update fan speed if changed
	if dutyCycle != m.state.FanSpeed {
		if err := m.setFanSpeedLocked(dutyCycle); err != nil {
			m.state.addWarning(fmt.Sprintf("Failed to update fan speed: %v", err))
		}
	}
}

// InitializeFanControl sets up PWM for fan control
func (m *Monitor) InitializeFanControl() error {
	if m.fanPin == "" {
		return nil // No fan control configured
	}

	err := m.gpio.ConfigurePWM(m.fanPin, nil, gpio.PWMConfig{
		Frequency: fanPWMFrequency,
		DutyCycle: fanSpeedLow,
	})
	if err != nil {
		return fmt.Errorf("failed to configure fan PWM: %w", err)
	}

	// Enable PWM output and set initial state
	if err := m.gpio.EnablePWM(m.fanPin); err != nil {
		return err
	}

	m.mux.Lock()
	m.state.FanSpeed = fanSpeedLow
	m.mux.Unlock()

	return nil
}

// setFanSpeedLocked controls fan speed using PWM - must be called with lock held
func (m *Monitor) setFanSpeedLocked(dutyCycle uint32) error {
	if m.fanPin == "" {
		return nil
	}

	// Clamp duty cycle to valid range
	if dutyCycle < fanSpeedLow {
		dutyCycle = fanSpeedLow
	}
	if dutyCycle > fanSpeedHigh {
		dutyCycle = fanSpeedHigh
	}

	if err := m.gpio.SetPWMDutyCycle(m.fanPin, dutyCycle); err != nil {
		return fmt.Errorf("failed to set fan PWM: %w", err)
	}
	m.state.FanSpeed = dutyCycle
	return nil
}

// setThrottlingLocked controls the throttling GPIO pin - must be called with lock held
func (m *Monitor) setThrottlingLocked(enabled bool) {
	if m.throttlePin == "" {
		return
	}

	if err := m.gpio.SetPinState(m.throttlePin, enabled); err != nil {
		m.state.addWarning(fmt.Sprintf("Failed to set throttling state: %v", err))
	}
	m.state.Throttled = enabled
}

// Close releases fan control resources
func (m *Monitor) Close() error {
	if m.fanPin != "" {
		if err := m.gpio.DisablePWM(m.fanPin); err != nil {
			return fmt.Errorf("failed to disable fan PWM: %w", err)
		}
	}
	return nil
}
