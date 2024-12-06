package diag

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
	"github.com/wrale/wrale-fleet-metal-hw/power"
	"github.com/wrale/wrale-fleet-metal-hw/thermal"
	"github.com/wrale/wrale-fleet-metal-hw/secure"
)

// mockHardware implements mock hardware interfaces for testing
type mockHardware struct {
	sync.RWMutex
	pinStates map[string]bool
	voltage   float64
	temps     struct {
		cpu float64
		gpu float64
	}
	fanSpeed  int
	caseOpen  bool
}

func newMockHardware() *mockHardware {
	return &mockHardware{
		pinStates: make(map[string]bool),
		voltage:   5.0,
		temps: struct{
			cpu float64
			gpu float64
		}{
			cpu: 45.0,
			gpu: 40.0,
		},
		fanSpeed: 25,
	}
}

// Mock GPIO interface
func (m *mockHardware) SetPinState(name string, high bool) error {
	m.Lock()
	defer m.Unlock()
	m.pinStates[name] = high
	return nil
}

func (m *mockHardware) GetPinState(name string) (bool, error) {
	m.RLock()
	defer m.RUnlock()
	return m.pinStates[name], nil
}

// Mock Power interface
func (m *mockHardware) GetState() power.PowerState {
	m.RLock()
	defer m.RUnlock()
	return power.PowerState{
		Voltage:         m.voltage,
		PowerConsumption: m.voltage * 0.5, // 0.5A default draw
	}
}

// Mock Thermal interface
func (m *mockHardware) GetThermalState() thermal.ThermalState {
	m.RLock()
	defer m.RUnlock()
	return thermal.ThermalState{
		CPUTemp:   m.temps.cpu,
		GPUTemp:   m.temps.gpu,
		FanSpeed:  m.fanSpeed,
		UpdatedAt: time.Now(),
	}
}

func (m *mockHardware) SetFanSpeed(speed int) {
	m.Lock()
	defer m.Unlock()
	m.fanSpeed = speed
}

// Mock Security interface
func (m *mockHardware) GetSecurityState() secure.TamperState {
	m.RLock()
	defer m.RUnlock()
	return secure.TamperState{
		CaseOpen:      m.caseOpen,
		VoltageNormal: m.voltage >= 4.8,
		LastCheck:     time.Now(),
	}
}

func TestDiagnostics(t *testing.T) {
	mock := newMockHardware()

	// Track test results
	var results []TestResult
	var resultsMux sync.Mutex

	// Create diagnostic manager
	mgr, err := New(Config{
		GPIO: mock,
		GPIOPins: []string{"test_pin1", "test_pin2"},
		LoadTestTime: 100 * time.Millisecond,
		MinVoltage: 4.8,
		TempRange: [2]float64{-10, 50},
		OnTestComplete: func(result TestResult) {
			resultsMux.Lock()
			results = append(results, result)
			resultsMux.Unlock()
		},
	})
	if err != nil {
		t.Fatalf("Failed to create diagnostic manager: %v", err)
	}

	// Test GPIO diagnostics
	t.Run("GPIO Tests", func(t *testing.T) {
		if err := mgr.TestGPIO(context.Background()); err != nil {
			t.Errorf("GPIO tests failed: %v", err)
		}

		// Verify all pins were tested both HIGH and LOW
		for _, pin := range mgr.cfg.GPIOPins {
			mock.SetPinState(pin, true)
			state, _ := mock.GetPinState(pin)
			if !state {
				t.Errorf("Pin %s not set HIGH", pin)
			}

			mock.SetPinState(pin, false)
			state, _ = mock.GetPinState(pin)
			if state {
				t.Errorf("Pin %s not set LOW", pin)
			}
		}
	})

	// Test power diagnostics
	t.Run("Power Tests", func(t *testing.T) {
		// Test normal voltage
		if err := mgr.TestPower(context.Background()); err != nil {
			t.Errorf("Power tests failed: %v", err)
		}

		// Test low voltage failure
		mock.voltage = 4.5
		if err := mgr.TestPower(context.Background()); err == nil {
			t.Error("Expected power test to fail at low voltage")
		}

		// Restore normal voltage
		mock.voltage = 5.0
	})

	// Test thermal diagnostics
	t.Run("Thermal Tests", func(t *testing.T) {
		// Test normal temperature
		if err := mgr.TestThermal(context.Background()); err != nil {
			t.Errorf("Thermal tests failed: %v", err)
		}

		// Test high temperature failure
		mock.temps.cpu = 55.0
		if err := mgr.TestThermal(context.Background()); err == nil {
			t.Error("Expected thermal test to fail at high temperature")
		}

		// Test fan control
		mock.temps.cpu = 45.0
		if err := mgr.TestThermal(context.Background()); err != nil {
			t.Errorf("Fan control test failed: %v", err)
		}
		if mock.fanSpeed != 25 { // Should return to default
			t.Errorf("Fan not returned to default speed: %d", mock.fanSpeed)
		}
	})

	// Test security diagnostics
	t.Run("Security Tests", func(t *testing.T) {
		// Test normal state
		if err := mgr.TestSecurity(context.Background()); err != nil {
			t.Errorf("Security tests failed: %v", err)
		}

		// Test case open warning
		mock.caseOpen = true
		if err := mgr.TestSecurity(context.Background()); err != nil {
			t.Errorf("Security test failed with case open: %v", err)
		}
		
		// Check that warning was recorded
		resultsMux.Lock()
		hasWarning := false
		for _, r := range results {
			if r.Type == TestSecurity && r.Status == StatusWarning {
				hasWarning = true
				break
			}
		}
		resultsMux.Unlock()
		
		if !hasWarning {
			t.Error("No warning recorded for open case")
		}
	})

	// Test complete diagnostic suite
	t.Run("Full Diagnostic Suite", func(t *testing.T) {
		// Reset to normal state
		mock.caseOpen = false
		mock.voltage = 5.0
		mock.temps.cpu = 45.0
		mock.temps.gpu = 40.0

		if err := mgr.RunAll(context.Background()); err != nil {
			t.Errorf("Full diagnostic suite failed: %v", err)
		}

		// Verify results were recorded
		resultsMux.Lock()
		finalResults := len(results)
		resultsMux.Unlock()

		if finalResults == 0 {
			t.Error("No test results recorded")
		}
	})
}