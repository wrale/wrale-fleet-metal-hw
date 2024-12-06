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
	t.Parallel()

	// Set overall test timeout
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	done := make(chan struct{})
	go func() {
		defer close(done)

		ctrl, err := New()
		if err != nil {
			t.Errorf("Failed to create GPIO controller: %v", err)
			return
		}
		defer ctrl.Close()

		pin := &mockPWMPin{}
		pinName := "test_pwm"

		// Test PWM configuration
		err = ctrl.ConfigurePWM(pinName, pin, PWMConfig{
			Frequency:  1000,
			DutyCycle: 50,
			Pull:      gpio.Float,
		})
		if err != nil {
			t.Errorf("Failed to configure PWM: %v", err)
			return
		}

		// Verify pull was configured
		if pin.Pull() != gpio.Float {
			t.Error("Pull not configured correctly")
			return
		}

		// Enable PWM
		if err := ctrl.EnablePWM(pinName); err != nil {
			t.Errorf("Failed to enable PWM: %v", err)
			return
		}

		// Let PWM run for several cycles
		time.Sleep(200 * time.Millisecond)
		initialHigh := pin.GetHighCount()

		// Change duty cycle
		if err := ctrl.SetPWMDutyCycle(pinName, 75); err != nil {
			t.Errorf("Failed to set duty cycle: %v", err)
			return
		}

		// Let it run again and verify high counts
		time.Sleep(200 * time.Millisecond)
		finalHigh := pin.GetHighCount()
		
		// With 75% duty cycle vs 50%, we should see roughly 50% more HIGH counts
		increase := float64(finalHigh - initialHigh) / float64(initialHigh)
		if increase < 0.3 { // Allow some margin for timing variations
			t.Errorf("Higher duty cycle (75%%) did not result in expected increase in HIGH outputs (got %.2f%% increase)", increase*100)
		}

		// Disable PWM
		if err := ctrl.DisablePWM(pinName); err != nil {
			t.Errorf("Failed to disable PWM: %v", err)
			return
		}

		// Pin should be low when disabled
		if pin.Read() != gpio.Low {
			t.Error("Pin not set LOW when PWM disabled")
			return
		}

		// High count should not increase after disable
		highCount := pin.GetHighCount()
		time.Sleep(50 * time.Millisecond)
		if pin.GetHighCount() != highCount {
			t.Error("Pin still toggling after PWM disabled")
		}
	}()

	select {
	case <-timer.C:
		t.Fatal("Test timeout after 2s")
	case <-done:
		// Test completed successfully
	}
}