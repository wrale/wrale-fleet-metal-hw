package power

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestLoadTesting(t *testing.T) {
	// Create mock GPIO controller
	mockGPIO := &mockGPIO{
		pinStates: make(map[string]bool),
	}

	// Create power manager with mock ADCs
	manager, err := New(Config{
		GPIO:       mockGPIO,
		PowerPins: map[PowerSource]string{
			MainPower: "main_power",
		},
		BatteryADCPath: t.TempDir() + "/battery",
		VoltageADCPath: t.TempDir() + "/voltage",
		CurrentADCPath: t.TempDir() + "/current",
	})
	if err != nil {
		t.Fatalf("Failed to create power manager: %v", err)
	}

	// Test successful load test
	t.Run("Successful Test", func(t *testing.T) {
		var progress []LoadTestProgress
		var progressMux sync.Mutex

		cfg := LoadTestConfig{
			TargetCurrent: 2.0, // 2A test
			Duration:      500 * time.Millisecond,
			MinVoltage:    4.8,
			MaxRipple:     0.2,
			OnProgress: func(p LoadTestProgress) {
				progressMux.Lock()
				progress = append(progress, p)
				progressMux.Unlock()
			},
		}

		ctx := context.Background()
		if err := manager.RunLoadTest(ctx, cfg); err != nil {
			t.Errorf("Load test failed: %v", err)
		}

		// Verify progress was reported
		if len(progress) == 0 {
			t.Error("No progress updates received")
		}

		// Verify load banks were enabled and disabled
		enabled := false
		for pin, state := range mockGPIO.pinStates {
			if pin == "load_bank_1" && state {
				enabled = true
			}
		}
		if !enabled {
			t.Error("Load banks not properly controlled")
		}
	})

	// Test voltage drop failure
	t.Run("Voltage Drop", func(t *testing.T) {
		cfg := LoadTestConfig{
			TargetCurrent: 2.0,
			Duration:      500 * time.Millisecond,
			MinVoltage:    5.2, // Set high to force failure
			MaxRipple:     0.2,
		}

		ctx := context.Background()
		if err := manager.RunLoadTest(ctx, cfg); err == nil {
			t.Error("Expected test to fail due to voltage drop")
		}
	})

	// Test ripple failure
	t.Run("Excessive Ripple", func(t *testing.T) {
		cfg := LoadTestConfig{
			TargetCurrent: 2.0,
			Duration:      500 * time.Millisecond,
			MinVoltage:    4.8,
			MaxRipple:     0.1, // Set low to force failure
		}

		ctx := context.Background()
		if err := manager.RunLoadTest(ctx, cfg); err == nil {
			t.Error("Expected test to fail due to excessive ripple")
		}
	})

	// Test invalid configurations
	t.Run("Invalid Config", func(t *testing.T) {
		invalidCfgs := []LoadTestConfig{
			{TargetCurrent: 0, Duration: time.Second, MinVoltage: 4.8},
			{TargetCurrent: 2.0, Duration: 0, MinVoltage: 4.8},
			{TargetCurrent: 2.0, Duration: time.Second, MinVoltage: 0},
		}

		for _, cfg := range invalidCfgs {
			if err := manager.RunLoadTest(context.Background(), cfg); err == nil {
				t.Error("Expected error for invalid config")
			}
		}
	})

	// Test context cancellation
	t.Run("Context Cancel", func(t *testing.T) {
		cfg := LoadTestConfig{
			TargetCurrent: 2.0,
			Duration:      5 * time.Second,
			MinVoltage:    4.8,
		}

		ctx, cancel := context.WithCancel(context.Background())
		errCh := make(chan error)

		go func() {
			errCh <- manager.RunLoadTest(ctx, cfg)
		}()

		// Cancel after short delay
		time.Sleep(100 * time.Millisecond)
		cancel()

		select {
		case err := <-errCh:
			if err != context.Canceled {
				t.Errorf("Expected context.Canceled, got: %v", err)
			}
		case <-time.After(time.Second):
			t.Error("Test did not cancel in time")
		}
	})
}