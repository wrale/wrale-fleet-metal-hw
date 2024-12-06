package power

import (
	"context"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

func TestPowerManager(t *testing.T) {
	gpioCtrl, err := gpio.New()
	if err != nil {
		t.Fatalf("Failed to create GPIO controller: %v", err)
	}

	manager, err := New(Config{
		GPIO:     gpioCtrl,
		PowerPins: map[PowerSource]string{
			MainPower:    "main_power",
			BatteryPower: "battery_power",
		},
		BatteryADCPath: "/dev/null",
		VoltageADCPath: "/dev/null",
		CurrentADCPath: "/dev/null",
		MonitorInterval: 100 * time.Millisecond,
	})

	if err != nil {
		t.Fatalf("Failed to create power manager: %v", err)
	}

	t.Run("Power Source Monitoring", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			errCh <- manager.Monitor(ctx)
		}()

		// Give monitor time to run
		select {
		case err := <-errCh:
			if err != nil && err != context.DeadlineExceeded {
				t.Errorf("Monitor failed: %v", err)
			}
		case <-time.After(300 * time.Millisecond):
			t.Error("Monitor did not complete in time")
		}

		// Check state is being updated
		state := manager.GetState()
		if state.UpdatedAt.IsZero() {
			t.Error("State not being updated")
		}
	})
}