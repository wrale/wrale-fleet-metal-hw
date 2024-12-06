package thermal

import (
	"sync"
	"testing"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

// mockFanController implements a mock GPIO controller for testing
type mockFanController struct {
	sync.Mutex
	pwmEnabled    bool
	dutyCycle     uint32
	throttleState bool
}

func NewMockGPIO() *gpio.Controller {
	ctrl, _ := gpio.New()
	return ctrl
}

func TestCooling(t *testing.T) {
	gpioCtrl := NewMockGPIO()
	
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
		err := monitor.InitializeFanControl()
		if err != nil {
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
		err := monitor.Close()
		if err != nil {
			t.Errorf("Failed to close monitor: %v", err)
		}
	})
}