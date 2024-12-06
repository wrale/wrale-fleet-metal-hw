package power

import (
	"context"
	"sync"
	"testing"
	"time"

	hw_gpio "github.com/wrale/wrale-fleet-metal-hw/gpio"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
)

// mockPin implements a basic GPIO pin for testing
type mockPin struct {
	sync.Mutex
	state bool
	pull  gpio.Pull
}

func (m *mockPin) String() string         { return "mock" }
func (m *mockPin) Halt() error            { return nil }
func (m *mockPin) Name() string           { return "MOCK" }
func (m *mockPin) Number() int            { return 0 }
func (m *mockPin) Function() string       { return "In/Out" }
func (m *mockPin) DefaultPull() gpio.Pull { return gpio.Float }
func (m *mockPin) In(pull gpio.Pull, edge gpio.Edge) error {
	m.Lock()
	defer m.Unlock()
	m.pull = pull
	return nil
}
func (m *mockPin) Read() gpio.Level {
	m.Lock()
	defer m.Unlock()
	if m.state {
		return gpio.High
	}
	return gpio.Low
}
func (m *mockPin) Out(l gpio.Level) error {
	m.Lock()
	defer m.Unlock()
	m.state = l == gpio.High
	return nil
}
func (m *mockPin) Pull() gpio.Pull                              { return m.pull }
func (m *mockPin) PWM(duty gpio.Duty, f physic.Frequency) error { return nil }
func (m *mockPin) WaitForEdge(timeout time.Duration) bool       { return true }

func TestPowerManager(t *testing.T) {
	gpioCtrl, err := hw_gpio.New(hw_gpio.WithSimulation())
	if err != nil {
		t.Fatalf("Failed to create GPIO controller: %v", err)
	}

	mainPin := &mockPin{}
	batteryPin := &mockPin{}

	if err := gpioCtrl.ConfigurePin("main_power", mainPin, gpio.PullUp); err != nil {
		t.Fatalf("Failed to configure main power pin: %v", err)
	}
	if err := gpioCtrl.ConfigurePin("battery_power", batteryPin, gpio.PullUp); err != nil {
		t.Fatalf("Failed to configure battery power pin: %v", err)
	}

	manager, err := New(Config{
		GPIO: gpioCtrl,
		PowerPins: map[PowerSource]string{
			MainPower:    "main_power",
			BatteryPower: "battery_power",
		},
		BatteryADCPath:  "/dev/null",
		VoltageADCPath:  "/dev/null",
		CurrentADCPath:  "/dev/null",
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
