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

	ctrl, err := New()
	if err != nil {
		t.Fatalf("Failed to create GPIO controller: %v", err)
	}
	defer ctrl.Close()

	pin := &mockPWMPin{}
	pinName := "test_pwm"

	// Configure PWM
	err = ctrl.ConfigurePWM(pinName, pin, PWMConfig{
		Frequency:  1000,
		DutyCycle: 50,
		Pull:      gpio.Float,
	})
	if err != nil {
		t.Fatalf("Failed to configure PWM: %v", err)
	}

	// Verify pull was configured
	if pin.Pull() != gpio.Float {
		t.Error("Pull not configured correctly")
	}

	// Enable PWM and test 50% duty cycle
	if err := ctrl.EnablePWM(pinName); err != nil {
		t.Fatalf("Failed to enable PWM: %v", err)
	}

	initTime := time.Now()
	time.Sleep(100 * time.Millisecond)
	initialHigh := pin.GetHighCount()
	initialElapsed := time.Since(initTime)

	// Test 75% duty cycle
	if err := ctrl.SetPWMDutyCycle(pinName, 75); err != nil {
		t.Fatalf("Failed to set duty cycle: %v", err)
	}

	finalTime := time.Now()
	time.Sleep(100 * time.Millisecond)
	finalHigh := pin.GetHighCount()
	finalElapsed := time.Since(finalTime)

	// Calculate and compare rates
	initialRate := float64(initialHigh) / initialElapsed.Seconds()
	finalRate := float64(finalHigh-initialHigh) / finalElapsed.Seconds()
	rateIncrease := (finalRate - initialRate) / initialRate * 100

	if rateIncrease < 25 { // Should see roughly 50% increase for 75% vs 50% duty cycle
		t.Errorf("Duty cycle change did not result in expected rate increase: got %.2f%%, want >= 25%%", rateIncrease)
	}

	// Test PWM disable
	if err := ctrl.DisablePWM(pinName); err != nil {
		t.Fatalf("Failed to disable PWM: %v", err)
	}

	// Pin should be low when disabled
	if pin.Read() != gpio.Low {
		t.Error("Pin not set LOW when PWM disabled")
	}

	// High count should not increase after disable
	highCount := pin.GetHighCount()
	time.Sleep(50 * time.Millisecond)
	if pin.GetHighCount() != highCount {
		t.Error("Pin still toggling after PWM disabled")
	}
}