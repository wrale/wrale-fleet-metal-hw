package thermal

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// updateThermalState reads current temperatures and updates cooling
func (m *Monitor) updateThermalState() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	// Read temperatures
	var warnings []string

	// Read CPU temperature
	if m.cpuTemp != "" {
		cpuTemp, err := m.readTemp(m.cpuTemp)
		if err != nil {
			return fmt.Errorf("failed to read CPU temperature: %w", err)
		}
		m.state.CPUTemp = cpuTemp

		// Check CPU temperature thresholds
		if cpuTemp >= cpuTempCritical {
			warnings = append(warnings, "CPU temperature critical")
			if m.onCritical != nil {
				m.onCritical(m.state)
			}
		} else if cpuTemp >= cpuTempWarning {
			warnings = append(warnings, "CPU temperature warning")
			if m.onWarning != nil {
				m.onWarning(m.state)
			}
		}
	}

	// Read GPU temperature
	if m.gpuTemp != "" {
		gpuTemp, err := m.readTemp(m.gpuTemp)
		if err != nil {
			return fmt.Errorf("failed to read GPU temperature: %w", err)
		}
		m.state.GPUTemp = gpuTemp

		// Check GPU temperature thresholds
		if gpuTemp >= gpuTempCritical {
			warnings = append(warnings, "GPU temperature critical")
			if m.onCritical != nil {
				m.onCritical(m.state)
			}
		} else if gpuTemp >= gpuTempWarning {
			warnings = append(warnings, "GPU temperature warning")
			if m.onWarning != nil {
				m.onWarning(m.state)
			}
		}
	}

	// Read ambient temperature
	if m.ambientTemp != "" {
		ambientTemp, err := m.readTemp(m.ambientTemp)
		if err != nil {
			return fmt.Errorf("failed to read ambient temperature: %w", err)
		}
		m.state.AmbientTemp = ambientTemp

		// Check ambient temperature thresholds
		if ambientTemp >= ambientCritical {
			warnings = append(warnings, "Ambient temperature critical")
			if m.onCritical != nil {
				m.onCritical(m.state)
			}
		} else if ambientTemp >= ambientWarning {
			warnings = append(warnings, "Ambient temperature warning")
			if m.onWarning != nil {
				m.onWarning(m.state)
			}
		}
	}

	// Update warnings
	m.state.Warnings = warnings

	// Determine required cooling
	m.updateCooling()

	m.state.UpdatedAt = time.Now()
	return nil
}

// readTemp reads a temperature value from sysfs
func (m *Monitor) readTemp(path string) (float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	// Convert raw value (usually in millicelsius)
	raw, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	if err != nil {
		return 0, err
	}

	// Convert to Celsius
	return raw / 1000.0, nil
}