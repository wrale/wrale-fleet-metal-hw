package thermal

import (
	"fmt"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

// Fan speed ranges and PWM settings
const (
	// Fan speed percentages
	fanSpeedLow    = 25
	fanSpeedMedium = 50
	fanSpeedHigh   = 100

	// PWM configuration
	fanPWMFrequency = 25000  // 25kHz standard for PC fans
)

// updateCooling adjusts cooling based on temperatures
func (m *Monitor) updateCooling() {
	// Determine maximum temperature
	maxTemp := m.state.CPUTemp
	if m.state.GPUTemp > maxTemp {
		maxTemp = m.state.GPUTemp
	}

	// Calculate fan speed based on temperature ranges
	var fanSpeed int
	switch {
	case maxTemp >= cpuTempCritical:
		fanSpeed = fanSpeedHigh
		m.setThrottling(true)
	case maxTemp >= cpuTempWarning:
		// Linear interpolation between medium and high speed
		tempRange := cpuTempCritical - cpuTempWarning
		tempAboveWarning := maxTemp - cpuTempWarning
		speedRange := fanSpeedHigh - fanSpeedMedium
		fanSpeed = fanSpeedMedium + int(float64(speedRange)*(tempAboveWarning/tempRange))
		m.setThrottling(false)
	case maxTemp >= (cpuTempWarning/2):
		// Linear interpolation between low and medium speed
		tempRange := cpuTempWarning - (cpuTempWarning/2)
		tempAboveMin := maxTemp - (cpuTempWarning/2)
		speedRange := fanSpeedMedium - fanSpeedLow
		fanSpeed = fanSpeedLow + int(float64(speedRange)*(tempAboveMin/tempRange))
		m.setThrottling(false)
	default:
		fanSpeed = fanSpeedLow
		m.setThrottling(false)
	}

	// Update fan speed if changed
	if fanSpeed != m.state.FanSpeed {
		m.setFanSpeed(fanSpeed)
		m.state.FanSpeed = fanSpeed
	}
}

// InitializeFanControl sets up PWM for fan control
func (m *Monitor) InitializeFanControl() error {
	if m.fanPin == "" {
		return nil // No fan control configured
	}

	err := m.gpio.ConfigurePWM(m.fanPin, nil, gpio.PWMConfig{
		Frequency:  fanPWMFrequency,
		DutyCycle: uint32(fanSpeedLow),
	})
	if err != nil {
		return fmt.Errorf("failed to configure fan PWM: %w", err)
	}

	// Enable PWM output
	return m.gpio.EnablePWM(m.fanPin)
}

// setFanSpeed controls fan speed using PWM
func (m *Monitor) setFanSpeed(speed int) {
	if m.fanPin == "" {
		return
	}

	// Clamp speed to valid range
	if speed < fanSpeedLow {
		speed = fanSpeedLow
	}
	if speed > fanSpeedHigh {
		speed = fanSpeedHigh
	}

	m.gpio.SetPWMDutyCycle(m.fanPin, uint32(speed))
}

// setThrottling controls the throttling GPIO pin
func (m *Monitor) setThrottling(enabled bool) {
	if m.throttlePin == "" {
		return
	}

	m.gpio.SetPinState(m.throttlePin, enabled)
	m.state.Throttled = enabled
}

// Close releases fan control resources
func (m *Monitor) Close() error {
	if m.fanPin != "" {
		if err := m.gpio.DisablePWM(m.fanPin); err != nil {
			return err
		}
	}
	return nil
}