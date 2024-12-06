package thermal

import (
	"sync"
	"testing"
	"time"

	hw_gpio "github.com/wrale/wrale-fleet-metal-hw/gpio"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
)

// mockFanPin implements a mock GPIO pin for fan control
type mockFanPin struct {
	sync.Mutex
	state     gpio.Level
	pull      gpio.Pull
	dutyCycle uint32
}

func (m *mockFanPin) String() string                   { return "mock_fan" }
func (m *mockFanPin) Halt() error                      { return nil }
func (m *mockFanPin) Name() string                     { return "MOCK_FAN" }
func (m *mockFanPin) Number() int                      { return 0 }
func (m *mockFanPin) Function() string                 { return "PWM" }
func (m *mockFanPin) DefaultPull() gpio.Pull           { return gpio.Float }
func (m *mockFanPin) PWM(duty gpio.Duty, f physic.Frequency) error { return nil }
func (m *mockFanPin) In(pull gpio.Pull, edge gpio.Edge) error {
	m.Lock()
	defer m.Unlock()
	m.pull = pull
	return nil
}
func (m *mockFanPin) Read() gpio.Level {
	m.Lock()
	defer m.Unlock()
	return m.state
}
func (m *mockFanPin) Out(l gpio.Level) error {
	m.Lock()
	defer m.Unlock()
	m.state = l
	if l == gpio.High {
		m.dutyCycle = 100
	} else {
		m.dutyCycle = 0
	}
	return nil
}
func (m *mockFanPin) Pull() gpio.Pull {
	m.Lock()
	defer m.Unlock()
	return m.pull
}
func (m *mockFanPin) WaitForEdge(timeout time.Duration) bool { return true }

// mockThrottlePin implements a mock GPIO pin for throttle control
type mockThrottlePin struct {
	sync.Mutex
	state bool
}

func (m *mockThrottlePin) String() string             { return "mock_throttle" }
func (m *mockThrottlePin) Halt() error                { return nil }
func (m *mockThrottlePin) Name() string               { return "MOCK_THROTTLE" }
func (m *mockThrottlePin) Number() int                { return 0 }
func (m *mockThrottlePin) Function() string           { return "In/Out" }
func (m *mockThrottlePin) DefaultPull() gpio.Pull     { return gpio.Float }
func (m *mockThrottlePin) In(pull gpio.Pull, edge gpio.Edge) error {
	m.Lock()
	defer m.Unlock()
	return nil
}
func (m *mockThrottlePin) Read() gpio.Level {
	m.Lock()
	defer m.Unlock()
	if m.state {
		return gpio.High
	}
	return gpio.Low
}
func (m *mockThrottlePin) Out(l gpio.Level) error {
	m.Lock()
	defer m.Unlock()
	m.state = l == gpio.High
	return nil
}
func (m *mockThrottlePin) Pull() gpio.Pull {
	m.Lock()
	defer m.Unlock()
	return gpio.Float
}
func (m *mockThrottlePin) PWM(duty gpio.Duty, f physic.Frequency) error { return nil }
func (m *mockThrottlePin) WaitForEdge(timeout time.Duration) bool { return true }

func TestCooling(t *testing.T) {
	gpioCtrl, err := hw_gpio.New()
	if err != nil {
		t.Fatalf("Failed to create GPIO controller: %v", err)
	}

	fanPin := &mockFanPin{}
	throttlePin := &mockThrottlePin{}

	// Configure pins before creating monitor
	if err := gpioCtrl.ConfigurePin("test_fan", fanPin, gpio.Float); err != nil {
		t.Fatalf("Failed to configure fan pin: %v", err)
	}
	if err := gpioCtrl.ConfigurePin("test_throttle", throttlePin, gpio.Float); err != nil {
		t.Fatalf("Failed to configure throttle pin: %v", err)
	}

	monitor := &Monitor{
		gpio:        gpioCtrl,
		fanPin:      "test_fan",
		throttlePin: "test_throttle",
		state: ThermalState{
			CPUTemp: 45.0,
			GPUTemp: 40.0,
		},
	}

	// Test fan initialization
	t.Run("Fan Initialization", func(t *testing.T) {
		if err := monitor.InitializeFanControl(); err != nil {
			t.Errorf("Failed to initialize fan control: %v", err)
		}
	})

	// Test temperature-based speed control
	t.Run("Temperature Response", func(t *testing.T) {
		// Test low temperature
		monitor.state.CPUTemp = 35.0
		monitor.updateCooling()
		if monitor.state.FanSpeed != fanSpeedLow {
			t.Errorf("Expected fan speed %d at low temp, got %d", fanSpeedLow, monitor.state.FanSpeed)
		}
		if monitor.state.Throttled {
			t.Error("Throttling enabled at low temperature")
		}

		// Test critical temperature
		monitor.state.CPUTemp = cpuTempCritical
		monitor.updateCooling()
		if monitor.state.FanSpeed != fanSpeedHigh {
			t.Error("Fan not at full speed at critical temperature")
		}
		if !monitor.state.Throttled {
			t.Error("Throttling not enabled at critical temperature")
		}
	})

	// Test cleanup
	t.Run("Cleanup", func(t *testing.T) {
		if err := monitor.Close(); err != nil {
			t.Errorf("Failed to close monitor: %v", err)
		}
	})
}