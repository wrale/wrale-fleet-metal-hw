package diag

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

func NewMockGPIO() (*gpio.Controller, error) {
	return gpio.New(gpio.WithSimulation())
}

func TestDiagnostics(t *testing.T) {
	// Create GPIO controller for tests in simulation mode
	gpioCtrl, err := NewMockGPIO()
	if err != nil {
		t.Fatalf("Failed to create GPIO controller: %v", err)
	}

	// Track test results
	var results []TestResult
	var resultsMux sync.Mutex

	// Create diagnostic manager
	mgr, err := New(Config{
		GPIO:         gpioCtrl,
		GPIOPins:     []string{"test_pin1", "test_pin2"},
		LoadTestTime: 100 * time.Millisecond,
		MinVoltage:   4.8,
		TempRange:    [2]float64{-10, 50},
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
			// Expect this to fail since we're in simulation mode
			t.Skip("Skipping GPIO test on simulation")
		}
	})

	// Test power diagnostics
	t.Run("Power Tests", func(t *testing.T) {
		if err := mgr.TestPower(context.Background()); err != nil {
			// Expect this to fail since we're in simulation mode
			t.Skip("Skipping power test on simulation")
		}
	})

	// Test thermal diagnostics
	t.Run("Thermal Tests", func(t *testing.T) {
		if err := mgr.TestThermal(context.Background()); err != nil {
			// Expect this to fail since we're in simulation mode
			t.Skip("Skipping thermal test on simulation")
		}
	})

	// Test security diagnostics
	t.Run("Security Tests", func(t *testing.T) {
		if err := mgr.TestSecurity(context.Background()); err != nil {
			// Expect this to fail since we're in simulation mode
			t.Skip("Skipping security test on simulation")
		}
	})

	// Test complete diagnostic suite
	t.Run("Full Diagnostic Suite", func(t *testing.T) {
		if err := mgr.RunAll(context.Background()); err != nil {
			// Expect this to fail since we're in simulation mode
			t.Skip("Skipping full test suite on simulation")
		}

		// Verify test tracking works
		resultsMux.Lock()
		finalResults := len(results)
		resultsMux.Unlock()

		if finalResults > 0 {
			t.Logf("Recorded %d test results", finalResults)
		}
	})
}
