package diag

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager handles hardware diagnostics and testing
type Manager struct {
	mux  sync.RWMutex
	cfg  Config
	
	// Test history
	results []TestResult
}

// New creates a new hardware diagnostics manager
func New(cfg Config) (*Manager, error) {
	if cfg.GPIO == nil {
		return nil, fmt.Errorf("GPIO controller required")
	}

	// Set defaults
	if cfg.Retries == 0 {
		cfg.Retries = 3
	}
	if cfg.LoadTestTime == 0 {
		cfg.LoadTestTime = 30 * time.Second
	}
	if cfg.MinVoltage == 0 {
		cfg.MinVoltage = 4.8 // 4.8V minimum for 5V system
	}
	if cfg.TempRange == [2]float64{} {
		cfg.TempRange = [2]float64{-10, 50} // -10°C to 50°C
	}

	return &Manager{
		cfg: cfg,
	}, nil
}

// TestGPIO performs GPIO pin diagnostics
func (m *Manager) TestGPIO(ctx context.Context) error {
	for _, pin := range m.cfg.GPIOPins {
		// Test output mode
		if err := m.cfg.GPIO.SetPinState(pin, true); err != nil {
			m.recordResult(TestResult{
				Type:        TestGPIO,
				Component:   pin,
				Status:      StatusFail,
				Description: "Failed to set pin HIGH",
				Error:       err,
				Timestamp:   time.Now(),
			})
			return fmt.Errorf("failed to set pin %s HIGH: %w", pin, err)
		}

		// Verify state
		state, err := m.cfg.GPIO.GetPinState(pin)
		if err != nil || !state {
			m.recordResult(TestResult{
				Type:        TestGPIO,
				Component:   pin,
				Status:      StatusFail,
				Description: "Pin readback mismatch",
				Error:       err,
				Timestamp:   time.Now(),
			})
			return fmt.Errorf("pin %s state mismatch", pin)
		}

		m.recordResult(TestResult{
			Type:        TestGPIO,
			Component:   pin,
			Status:      StatusPass,
			Description: "GPIO pin functional",
			Timestamp:   time.Now(),
		})
	}

	return nil
}

// TestPower performs power subsystem diagnostics
func (m *Manager) TestPower(ctx context.Context) error {
	if m.cfg.Power == nil {
		return fmt.Errorf("power manager not configured")
	}

	// Test power stability
	state := m.cfg.Power.GetState()
	if state.Voltage < m.cfg.MinVoltage {
		m.recordResult(TestResult{
			Type:        TestPower,
			Component:   "voltage",
			Status:      StatusFail,
			Reading:     state.Voltage,
			Expected:    m.cfg.MinVoltage,
			Description: "Voltage below minimum",
			Timestamp:   time.Now(),
		})
		return fmt.Errorf("voltage %v below minimum %v", state.Voltage, m.cfg.MinVoltage)
	}

	m.recordResult(TestResult{
		Type:        TestPower,
		Component:   "power_system",
		Status:      StatusPass,
		Description: "Power system functional",
		Timestamp:   time.Now(),
	})

	return nil
}

// TestThermal performs thermal subsystem diagnostics
func (m *Manager) TestThermal(ctx context.Context) error {
	if m.cfg.Thermal == nil {
		return fmt.Errorf("thermal monitor not configured")
	}

	state := m.cfg.Thermal.GetState()

	// Verify temperature readings
	if state.CPUTemp < m.cfg.TempRange[0] || state.CPUTemp > m.cfg.TempRange[1] {
		m.recordResult(TestResult{
			Type:        TestThermal,
			Component:   "cpu_temp",
			Status:      StatusFail,
			Reading:     state.CPUTemp,
			Description: "CPU temperature out of range",
			Timestamp:   time.Now(),
		})
		return fmt.Errorf("CPU temp %v outside range %v-%v", state.CPUTemp, 
			m.cfg.TempRange[0], m.cfg.TempRange[1])
	}

	if state.GPUTemp < m.cfg.TempRange[0] || state.GPUTemp > m.cfg.TempRange[1] {
		m.recordResult(TestResult{
			Type:        TestThermal,
			Component:   "gpu_temp",
			Status:      StatusFail,
			Reading:     state.GPUTemp,
			Description: "GPU temperature out of range",
			Timestamp:   time.Now(),
		})
		return fmt.Errorf("GPU temp %v outside range %v-%v", state.GPUTemp,
			m.cfg.TempRange[0], m.cfg.TempRange[1])
	}

	// Set fan to low speed for test
	if err := m.cfg.Thermal.SetFanSpeed(25); err != nil {
		m.recordResult(TestResult{
			Type:        TestThermal,
			Component:   "fan",
			Status:      StatusFail,
			Description: "Failed to control fan speed",
			Error:       err,
			Timestamp:   time.Now(),
		})
		return fmt.Errorf("failed to control fan: %w", err)
	}

	m.recordResult(TestResult{
		Type:        TestThermal,
		Component:   "thermal_system",
		Status:      StatusPass,
		Description: "Thermal system functional",
		Timestamp:   time.Now(),
	})

	return nil
}

// TestSecurity performs security subsystem diagnostics
func (m *Manager) TestSecurity(ctx context.Context) error {
	if m.cfg.Security == nil {
		return fmt.Errorf("security manager not configured")
	}

	state := m.cfg.Security.GetState()

	// Verify security sensors respond
	if state.CaseOpen {
		m.recordResult(TestResult{
			Type:        TestSecurity,
			Component:   "case_sensor",
			Status:      StatusWarning,
			Description: "Case open detected",
			Timestamp:   time.Now(),
		})
	}

	if state.MotionDetected {
		m.recordResult(TestResult{
			Type:        TestSecurity,
			Component:   "motion_sensor",
			Status:      StatusWarning,
			Description: "Motion detected",
			Timestamp:   time.Now(),
		})
	}

	if !state.VoltageNormal {
		m.recordResult(TestResult{
			Type:        TestSecurity,
			Component:   "voltage_monitor",
			Status:      StatusFail,
			Description: "Abnormal voltage detected",
			Timestamp:   time.Now(),
		})
		return fmt.Errorf("security voltage monitor shows abnormal state")
	}

	m.recordResult(TestResult{
		Type:        TestSecurity,
		Component:   "security_system",
		Status:      StatusPass,
		Description: "Security system functional",
		Timestamp:   time.Now(),
	})

	return nil
}

// RunAll performs a complete hardware diagnostic suite
func (m *Manager) RunAll(ctx context.Context) error {
	tests := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"GPIO", m.TestGPIO},
		{"Power", m.TestPower},
		{"Thermal", m.TestThermal},
		{"Security", m.TestSecurity},
	}

	for _, test := range tests {
		for retry := 0; retry < m.cfg.Retries; retry++ {
			err := test.fn(ctx)
			if err == nil {
				break
			}
			if retry == m.cfg.Retries-1 {
				return fmt.Errorf("%s tests failed after %d retries: %w", 
					test.name, m.cfg.Retries, err)
			}
			time.Sleep(time.Second) // Wait between retries
		}
	}

	return nil
}

// GetResults returns all test results
func (m *Manager) GetResults() []TestResult {
	m.mux.RLock()
	defer m.mux.RUnlock()
	
	results := make([]TestResult, len(m.results))
	copy(results, m.results)
	return results
}

// recordResult stores a test result and notifies callback if configured
func (m *Manager) recordResult(result TestResult) {
	m.mux.Lock()
	m.results = append(m.results, result)
	m.mux.Unlock()

	if m.cfg.OnTestComplete != nil {
		m.cfg.OnTestComplete(result)
	}
}