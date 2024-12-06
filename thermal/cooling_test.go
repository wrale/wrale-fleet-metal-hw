package thermal

import (
	"testing"
	"sync"
	"time"
	
	"periph.io/x/conn/v3/gpio"
)

// mockFanController implements a mock GPIO controller for testing fan control
type mockFanController struct {
	sync.Mutex
	pwmEnabled    bool
	dutyCycle     uint32
	throttleState bool
}

func (m *mockFanController) ConfigurePWM(name string, pin gpio.PinIO, cfg gpio.PWMConfig) error {
	m.Lock()
	defer m.Unlock()
	return nil
}

func (m *mockFanController) EnablePWM(name string) error {
	m.Lock()
	defer m.Unlock()
	m.pwmEnabled = true
	return nil
}

func (m *mockFanController) DisablePWM(name string) error {
	m.Lock()
	defer m.Unlock()
	m.pwmEnabled = false
	return nil
}

func (m *mockFanController) SetPWMDutyCycle(name string, duty uint32) error {
	m.Lock()
	defer m.Unlock()
	m.dutyCycle = duty
	return nil
}

func (m *mockFanController) SetPinState(name string, high bool) error {
	m.Lock()
	defer m.Unlock()
	m.throttleState = high
	return nil
}

func TestCooling(t *testing.T) {
	mockGPIO := &mockFanController{}
	
	monitor := &Monitor{
		gpio:        mockGPIO,
		fanPin:      "fan",
		throttlePin: "throttle",
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
		if !mockGPIO.pwmEnabled {
			t.Error("PWM not enabled after initialization")
		}
	})

	// Test temperature-based speed control
	t.Run("Temperature Response", func(t *testing.T) {
		// Test low temperature
		monitor.state.CPUTemp = 35.0
		monitor.updateCooling()
		if mockGPIO.dutyCycle != uint32(fanSpeedLow) {
			t.Errorf("Expected fan speed %d at low temp, got %d", fanSpeedLow, mockGPIO.dutyCycle)
		}
		if mockGPIO.throttleState {
			t.Error("Throttling enabled at low temperature")
		}

		// Test warning temperature
		monitor.state.CPUTemp = cpuTempWarning
		monitor.updateCooling()
		if mockGPIO.dutyCycle < uint32(fanSpeedMedium) {
			t.Error("Fan speed not increased at warning temperature")
		}

		// Test critical temperature
		monitor.state.CPUTemp = cpuTempCritical
		monitor.updateCooling()
		if mockGPIO.dutyCycle != uint32(fanSpeedHigh) {
			t.Error("Fan not at full speed at critical temperature")
		}
		if !mockGPIO.throttleState {
			t.Error("Throttling not enabled at critical temperature")
		}
	})

	// Test cleanup
	t.Run("Cleanup", func(t *testing.T) {
		err := monitor.Close()
		if err != nil {
			t.Errorf("Failed to close monitor: %v", err)
		}
		if mockGPIO.pwmEnabled {
			t.Error("PWM still enabled after close")
		}
	})
}