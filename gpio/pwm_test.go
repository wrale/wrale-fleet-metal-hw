package gpio

import (
	"sync"
	"testing"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
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
func (m *mockPWMPin) DefaultPull() gpio.Pull           { return gpio.Float }
func (m *mockPWMPin) PWM(duty gpio.Duty, f physic.Frequency) error { return nil }
func (m *mockPWMPin) Pull() gpio.Pull { return m.pull }
func (m *mockPWMPin) WaitForEdge(timeout time.Duration) bool { return true }

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
	defer ctrl.Close()

	pin := &mockPWMPin{}
	pinName := "test_pwm"

	// Test PWM configuration
	t.Run("Configure PWM", func(t *testing.T) {
		err := ctrl.ConfigurePWM(pinName, pin, PWMConfig{
			Frequency:  1000,
			DutyCycle: 50,
			Pull:      gpio.Float,
		})
		if err != nil {
			t.Errorf("Failed to configure PWM: %v", err)
		}

		// Verify pull was configured
		if pin.Pull() != gpio.Float {
			t.Error("Pull not configured correctly")
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
}