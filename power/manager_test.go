package power

import (
	"context"
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

// setupTestADC creates a mock ADC sysfs file
func setupTestADC(t *testing.T, value string) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "adc_value")
	if err := os.WriteFile(path, []byte(value), 0644); err != nil {
		t.Fatalf("Failed to create test ADC file: %v", err)
	}
	return path
}

func TestPowerManager(t *testing.T) {
	// Setup mock GPIO
	mockGPIO := newMockGPIO()

	// Setup mock ADC files
	batteryADC := setupTestADC(t, "75.5") // 75.5% battery
	voltageADC := setupTestADC(t, "5.1")  // 5.1V
	currentADC := setupTestADC(t, "0.5")  // 0.5A

	// Setup power pins
	powerPins := map[PowerSource]string{
		MainPower:    "main_power",
		BatteryPower: "battery_power",
		SolarPower:   "solar_power",
	}

	// Create power manager
	var criticalPowerDetected bool
	manager, err := New(Config{
		GPIO:            mockGPIO,
		MonitorInterval: 100 * time.Millisecond,
		PowerPins:       powerPins,
		BatteryADCPath:  batteryADC,
		VoltageADCPath:  voltageADC,
		CurrentADCPath:  currentADC,
		OnPowerCritical: func(state PowerState) {
			criticalPowerDetected = true
		},
	})

	if err != nil {
		t.Fatalf("Failed to create power manager: %v", err)
	}

	// Start monitoring
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go func() {
		if err := manager.Monitor(ctx); err != nil && err != context.DeadlineExceeded {
			t.Errorf("Monitor failed: %v", err)
		}
	}()

	// Test normal power state
	t.Run("Normal Power State", func(t *testing.T) {
		// Set main power available
		mockGPIO.SetPinState("main_power", true)
		time.Sleep(200 * time.Millisecond)

		state := manager.GetState()
		if !state.AvailablePower[MainPower] {
			t.Error("Main power should be available")
		}
		if state.CurrentSource != MainPower {
			t.Error("Current source should be main power")
		}
		if state.BatteryLevel != 75.5 {
			t.Errorf("Expected battery level 75.5, got %f", state.BatteryLevel)
		}
		if state.Voltage != 5.1 {
			t.Errorf("Expected voltage 5.1, got %f", state.Voltage)
		}
		if state.PowerConsumption != 5.1*0.5 {
			t.Errorf("Expected power consumption %f, got %f", 5.1*0.5, state.PowerConsumption)
		}
	})

	// Test power failure scenario
	t.Run("Power Failure", func(t *testing.T) {
		// Remove main power
		mockGPIO.SetPinState("main_power", false)
		// Set battery power available but critical
		mockGPIO.SetPinState("battery_power", true)
		// Update battery ADC to critical level
		os.WriteFile(batteryADC, []byte("5.0"), 0644)
		
		time.Sleep(200 * time.Millisecond)

		state := manager.GetState()
		if state.AvailablePower[MainPower] {
			t.Error("Main power should not be available")
		}
		if state.CurrentSource != BatteryPower {
			t.Error("Current source should be battery power")
		}
		if !criticalPowerDetected {
			t.Error("Critical power condition not detected")
		}
	})

	// Test solar power scenario
	t.Run("Solar Power", func(t *testing.T) {
		// Set solar power available
		mockGPIO.SetPinState("solar_power", true)
		time.Sleep(200 * time.Millisecond)

		state := manager.GetState()
		if !state.AvailablePower[SolarPower] {
			t.Error("Solar power should be available")
		}
	})
}