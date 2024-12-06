package thermal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// mockGPIO implements a mock GPIO controller for testing
type mockGPIO struct {
	mux       sync.RWMutex
	pinStates map[string]bool
}

func newMockGPIO() *mockGPIO {
	return &mockGPIO{
		pinStates: make(map[string]bool),
	}
}

func (m *mockGPIO) ConfigurePin(name string, _ interface{}) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.pinStates[name] = false
	return nil
}

func (m *mockGPIO) SetPinState(name string, high bool) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.pinStates[name] = high
	return nil
}

func (m *mockGPIO) GetPinState(name string) (bool, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.pinStates[name], nil
}

// setupTempSensor creates a mock temperature sensor file
func setupTempSensor(t *testing.T, tempC float64) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "temp")
	
	// Write temperature in millicelsius (standard sysfs format)
	tempMC := int(tempC * 1000)
	if err := os.WriteFile(path, []byte(fmt.Sprintf("%d\n", tempMC)), 0644); err != nil {
		t.Fatalf("Failed to create temp sensor file: %v", err)
	}
	return path
}

func TestThermalMonitor(t *testing.T) {
	// Setup mock GPIO
	mockGPIO := newMockGPIO()

	// Setup mock temperature sensors
	cpuTemp := setupTempSensor(t, 45.0)    // 45°C - normal
	gpuTemp := setupTempSensor(t, 75.0)    // 75°C - warning
	ambientTemp := setupTempSensor(t, 30.0) // 30°C - normal

	// Track callbacks
	var warningCalled, criticalCalled bool
	var warningState, criticalState ThermalState

	// Create monitor
	monitor, err := New(Config{
		GPIO:            mockGPIO,
		MonitorInterval: 100 * time.Millisecond,
		CPUTempPath:     cpuTemp,
		GPUTempPath:     gpuTemp,
		AmbientTempPath: ambientTemp,
		FanControlPin:   "fan_control",
		ThrottlePin:     "throttle_control",
		OnWarning: func(state ThermalState) {
			warningCalled = true
			warningState = state
		},
		OnCritical: func(state ThermalState) {
			criticalCalled = true
			criticalState = state
		},
	})

	if err != nil {
		t.Fatalf("Failed to create thermal monitor: %v", err)
	}

	// Start monitoring
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go func() {
		if err := monitor.Monitor(ctx); err != nil && err != context.DeadlineExceeded {
			t.Errorf("Monitor failed: %v", err)
		}
	}()

	// Test initial conditions
	t.Run("Initial State", func(t *testing.T) {
		time.Sleep(200 * time.Millisecond)
		state := monitor.GetState()

		if state.CPUTemp != 45.0 {
			t.Errorf("Expected CPU temp 45.0, got %f", state.CPUTemp)
		}
		if state.GPUTemp != 75.0 {
			t.Errorf("Expected GPU temp 75.0, got %f", state.GPUTemp)
		}
		if state.AmbientTemp != 30.0 {
			t.Errorf("Expected ambient temp 30.0, got %f", state.AmbientTemp)
		}

		// Should be in warning state due to GPU temp
		if !warningCalled {
			t.Error("Warning callback not triggered for high GPU temp")
		}
		if len(state.Warnings) == 0 {
			t.Error("Expected warnings for high GPU temp")
		}

		// Fan should be at medium speed
		if state.FanSpeed != fanSpeedMedium {
			t.Errorf("Expected fan speed %d, got %d", fanSpeedMedium, state.FanSpeed)
		}
	})

	// Test critical temperature
	t.Run("Critical Temperature", func(t *testing.T) {
		// Set CPU to critical temperature
		os.WriteFile(cpuTemp, []byte("85000\n"), 0644) // 85°C
		time.Sleep(200 * time.Millisecond)

		state := monitor.GetState()
		if !criticalCalled {
			t.Error("Critical callback not triggered")
		}
		if !state.Throttled {
			t.Error("Throttling not enabled for critical temperature")
		}
		if state.FanSpeed != fanSpeedHigh {
			t.Error("Fan not at full speed for critical temperature")
		}
	})

	// Test cooling recovery
	t.Run("Cooling Recovery", func(t *testing.T) {
		// Set all temperatures to normal
		os.WriteFile(cpuTemp, []byte("35000\n"), 0644)    // 35°C
		os.WriteFile(gpuTemp, []byte("40000\n"), 0644)    // 40°C
		os.WriteFile(ambientTemp, []byte("25000\n"), 0644) // 25°C
		time.Sleep(200 * time.Millisecond)

		state := monitor.GetState()
		if state.Throttled {
			t.Error("Throttling still enabled after temperature recovery")
		}
		if state.FanSpeed != fanSpeedLow {
			t.Error("Fan speed not reduced after temperature recovery")
		}
		if len(state.Warnings) > 0 {
			t.Error("Warnings not cleared after temperature recovery")
		}
	})
}