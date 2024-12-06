package gpio

import (
	"sync"
	"testing"
	"time"

	"periph.io/x/conn/v3/gpio"
)

// mockPWMPin implements a mock GPIO pin with PWM support
type mockPWMPin struct {
	sync.Mutex
	state     gpio.Level
	pull      gpio.Pull
	highCount int
	lowCount  int
}

func (m *mockPWMPin) String() string                   { return "mock_pwm" }
func (m *mockPWMPin) Halt() error                      { return nil }
func (m *mockPWMPin) Name() string                     { return "MOCK_PWM" }
func (m *mockPWMPin) Number() int                      { return 0 }
func (m *mockPWMPin) Function() string                 { return "PWM" }

func (m *mockPWMPin) In(pull gpio.Pull, edge gpio.Edge) error {
	m.Lock()
	defer m.Unlock()
	m.pull = pull
	return nil
}

func (m *mockPWMPin) Read() gpio.Level {
	m.Lock()
	defer m.Unlock()
	return m.state
}

func (m *mockPWMPin) Out(l gpio.Level) error {
	m.Lock()
	defer m.Unlock()
	m.state = l
	if l == gpio.High {
		m.highCount++
	} else {
		m.lowCount++
	}
	return nil
}

func TestPWM(t *testing.T) {
	ctrl, err := New()
	if err != nil {
		t.Fatalf("Failed to create GPIO controller: %v", err)
	}

	pin := &mockPWMPin{}
	pinName := "test_pwm"

	// Test PWM configuration
	t.Run("Configure PWM", func(t *testing.T) {
		cfg := PWMConfig{
			Frequency:  1000,
			DutyCycle: 50,
			Pull:      gpio.PullUp,
		}

		err := ctrl.ConfigurePWM(pinName, pin, cfg)
		if err != nil {
			t.Errorf("Failed to configure PWM: %v", err)
		}

		// Verify pull-up was configured
		if pin.pull != gpio.PullUp {
			t.Error("Pull-up not configured correctly")
		}
	})

	// Test duty cycle changes
	t.Run("Duty Cycle", func(t *testing.T) {
		// Enable PWM
		if err := ctrl.EnablePWM(pinName); err != nil {
			t.Errorf("Failed to enable PWM: %v", err)
		}

		// Let PWM run for a bit
		time.Sleep(50 * time.Millisecond)
		initialHigh := pin.highCount

		// Change duty cycle
		if err := ctrl.SetPWMDutyCycle(pinName, 75); err != nil {
			t.Errorf("Failed to set duty cycle: %v", err)
		}

		// Let it run again and verify more high counts with higher duty cycle
		time.Sleep(50 * time.Millisecond)
		if pin.highCount-initialHigh <= initialHigh {
			t.Error("Higher duty cycle did not result in more HIGH outputs")
		}
	})

	// Test PWM disable
	t.Run("Disable PWM", func(t *testing.T) {
		if err := ctrl.DisablePWM(pinName); err != nil {
			t.Errorf("Failed to disable PWM: %v", err)
		}

		// Pin should be low when disabled
		if pin.Read() != gpio.Low {
			t.Error("Pin not set LOW when PWM disabled")
		}

		// High count should not increase after disable
		highCount := pin.highCount
		time.Sleep(50 * time.Millisecond)
		if pin.highCount != highCount {
			t.Error("Pin still toggling after PWM disabled")
		}
	})

	// Test invalid configurations
	t.Run("Invalid Config", func(t *testing.T) {
		// Test duty cycle > 100
		if err := ctrl.SetPWMDutyCycle(pinName, 101); err == nil {
			t.Error("Expected error for duty cycle > 100")
		}

		// Test non-existent pin
		if err := ctrl.EnablePWM("nonexistent"); err == nil {
			t.Error("Expected error for non-existent PWM pin")
		}

		// Test invalid frequency
		cfg := PWMConfig{
			Frequency: 0,
			DutyCycle: 50,
		}
		if err := ctrl.ConfigurePWM("bad_freq", pin, cfg); err == nil {
			t.Error("Expected error for zero frequency")
		}
	})

	// Cleanup
	if err := ctrl.Close(); err != nil {
		t.Errorf("Failed to close controller: %v", err)
	}
}