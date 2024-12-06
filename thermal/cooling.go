package thermal

// updateCooling adjusts cooling based on temperatures
func (m *Monitor) updateCooling() {
	// Determine maximum temperature
	maxTemp := m.state.CPUTemp
	if m.state.GPUTemp > maxTemp {
		maxTemp = m.state.GPUTemp
	}

	// Set fan speed based on temperature
	var fanSpeed int
	switch {
	case maxTemp >= cpuTempCritical:
		fanSpeed = fanSpeedHigh
		m.setThrottling(true)
	case maxTemp >= cpuTempWarning:
		fanSpeed = fanSpeedMedium
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

// setFanSpeed controls the fan GPIO pin
func (m *Monitor) setFanSpeed(speed int) {
	if m.fanPin == "" {
		return
	}

	// TODO: Implement PWM control for variable speed
	// For now, just on/off based on threshold
	m.gpio.SetPinState(m.fanPin, speed > fanSpeedLow)
}

// setThrottling controls the throttling GPIO pin
func (m *Monitor) setThrottling(enabled bool) {
	if m.throttlePin == "" {
		return
	}

	m.gpio.SetPinState(m.throttlePin, enabled)
	m.state.Throttled = enabled
}