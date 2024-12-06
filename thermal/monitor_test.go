package thermal

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

type mockGPIO struct {
	sync.RWMutex
	pinStates map[string]bool
}

func newMockGPIO() *gpio.Controller {
	ctrl, _ := gpio.New()
	return ctrl
}

func TestMonitor(t *testing.T) {
	gpioCtrl := newMockGPIO()
	
	monitor := &Monitor{
		gpio:        gpioCtrl,
		fanPin:      "test_fan",
		throttlePin: "test_throttle",
		cpuTemp:     "/tmp/cpu_temp",
		gpuTemp:     "/tmp/gpu_temp",
		state: ThermalState{
			CPUTemp: 45.0,
			GPUTemp: 40.0,
		},
	}

	t.Run("Monitor Operation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		var monitorErr error
		go func() {
			monitorErr = monitor.Monitor(ctx)
		}()

		<-ctx.Done()
		if monitorErr != nil && monitorErr != context.DeadlineExceeded {
			t.Errorf("Monitor failed: %v", monitorErr)
		}
	})
}