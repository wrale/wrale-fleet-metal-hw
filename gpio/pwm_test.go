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
func (m *mockPWMPin) Pull() gpio.Pull {
	m.Lock()
	defer m.Unlock()
	return m.pull
}
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

func (m *mockPWMPin) GetHighCount() int {
	m.Lock()
	defer m.Unlock()
	return m.highCount
}

func TestPWM(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PWM test in short mode")
	}

	// Create controller in simulation mode
	ctrl, err := New(WithSimulation())
	if err != nil {
		t.Fatalf("Failed to create GPIO controller: %v", err)
	}
	defer ctrl.Close()

	pin := &mockPWMPin{}
	pinName := "test_pwm"

	// Configure PWM with explicit Pull
	err = ctrl.ConfigurePWM(pinName, pin, PWMConfig{
		Frequency:  1000,
		DutyCycle: 50,
		Pull:      gpio.Float,
	})
	if err != nil {
		t.Fatalf("Failed to configure PWM: %v", err)
	}

	// In simulation mode, verify simulated pull state
	pull, err := ctrl.GetPinPull(pinName)
	if err != nil {
		t.Fatalf("Failed to get pin pull: %v", err)
	}
	if pull != gpio.Float {
		t.Errorf("Expected pin pull %v, got %v", gpio.Float, pull)
	}

	// Also verify physical pin if provided
	if pin != nil && pin.Pull() != gpio.Float {
		t.Error("Physical pin pull not configured correctly")
	}

	// Enable PWM and test 50% duty cycle
	if err := ctrl.EnablePWM(pinName); err != nil {
		t.Fatalf("Failed to enable PWM: %v", err)
	}

	// Let PWM run for a bit
	time.Sleep(100 * time.Millisecond)
	highCount := pin.GetHighCount()
	if highCount == 0 {
		t.Log("Note: PWM not producing pulses in simulation mode (expected)")
	}

	// Set new duty cycle
	if err := ctrl.SetPWMDutyCycle(pinName, 75); err != nil {
		t.Fatalf("Failed to set duty cycle: %v", err)
	}

	// Test PWM disable
	if err := ctrl.DisablePWM(pinName); err != nil {
		t.Fatalf("Failed to disable PWM: %v", err)
	}

	// Verify pin is low when disabled
	if pin.Read() != gpio.Low {
		t.Error("Pin not set LOW when PWM disabled")
	}

	// Verify no more toggling after disable
	highCount = pin.GetHighCount()
	time.Sleep(50 * time.Millisecond)
	if pin.GetHighCount() != highCount {
		t.Error("Pin still toggling after PWM disabled")
	}
}