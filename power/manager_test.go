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

		go func() {
			if err := manager.Monitor(ctx); err != nil && err != context.DeadlineExceeded {
				t.Errorf("Monitor failed: %v", err)
			}
		}()

		// Give monitor time to run
		time.Sleep(150 * time.Millisecond)

		// Check state is being updated
		state := manager.GetState()
		if state.UpdatedAt.IsZero() {
			t.Error("State not being updated")
		}
	})
}