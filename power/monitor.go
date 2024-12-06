package power

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Monitor starts monitoring power state in the background
func (m *Manager) Monitor(ctx context.Context) error {
	ticker := time.NewTicker(m.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.updatePowerState(); err != nil {
				return fmt.Errorf("failed to update power state: %w", err)
			}
		}
	}
}

// updatePowerState reads current power status from hardware
func (m *Manager) updatePowerState() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	// Check power sources
	for source, pin := range m.powerPins {
		available, err := m.gpio.GetPinState(pin)
		if err != nil {
			return fmt.Errorf("failed to check power source %s: %w", source, err)
		}
		m.state.AvailablePower[source] = available

		// Update current source if available
		if available {
			m.state.CurrentSource = source
		}
	}

	// Read battery level
	if m.batteryADC != "" {
		level, err := m.readADC(m.batteryADC)
		if err != nil {
			return fmt.Errorf("failed to read battery level: %w", err)
		}
		m.state.BatteryLevel = level
	}

	// Read voltage
	if m.voltageADC != "" {
		voltage, err := m.readADC(m.voltageADC)
		if err != nil {
			return fmt.Errorf("failed to read voltage: %w", err)
		}
		m.state.Voltage = voltage
	}

	// Read current consumption
	if m.currentADC != "" {
		current, err := m.readADC(m.currentADC)
		if err != nil {
			return fmt.Errorf("failed to read current: %w", err)
		}
		// Convert to power consumption (P = V * I)
		m.state.PowerConsumption = current * m.state.Voltage
	}

	// Update charging status - charging if main power available and battery not full
	m.state.Charging = m.state.AvailablePower[MainPower] && m.state.BatteryLevel < 100

	// Check for critical power conditions
	if !m.shutdownInitiated && m.isPowerCritical() {
		m.handleCriticalPower()
	}

	m.state.UpdatedAt = time.Now()
	return nil
}

// readADC reads a value from an ADC via sysfs
func (m *Manager) readADC(path string) (float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	// Convert raw ADC value
	raw, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	if err != nil {
		return 0, err
	}

	// TODO: Apply proper scaling based on ADC specs
	return raw, nil
}

// isPowerCritical checks if power state is critical
func (m *Manager) isPowerCritical() bool {
	// Critical if on battery and level is low
	if m.state.CurrentSource == BatteryPower && m.state.BatteryLevel <= criticalBatteryLevel {
		return true
	}

	// Critical if voltage is too low
	if m.state.Voltage <= criticalVoltage {
		return true
	}

	// Critical if no power sources available
	noAvailablePower := true
	for _, available := range m.state.AvailablePower {
		if available {
			noAvailablePower = false
			break
		}
	}

	return noAvailablePower
}

// handleCriticalPower handles critical power conditions
func (m *Manager) handleCriticalPower() {
	m.shutdownInitiated = true

	// Notify via callback if configured
	if m.onPowerCritical != nil {
		m.onPowerCritical(m.state)
	}
}