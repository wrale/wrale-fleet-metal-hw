package thermal

import (
	"context"
	"fmt"
	"sync"
	"time"

	pkggpio "github.com/wrale/wrale-fleet-metal-hw/gpio"
	"periph.io/x/conn/v3/gpio"
)

// Monitor handles thermal monitoring and control
type Monitor struct {
	mux   sync.RWMutex
	state ThermalState

	// Hardware interface
	gpio       *pkggpio.Controller
	fanPin     string
	throttlePin string

	// Temperature paths
	cpuTemp     string
	gpuTemp     string
	ambientTemp string

	// Configuration
	monitorInterval time.Duration
	onWarning       func(ThermalState)
	onCritical      func(ThermalState)

	// Internal state
	shutdownInitiated bool
}

// New creates a new thermal monitor
func New(cfg Config) (*Monitor, error) {
	if cfg.GPIO == nil {
		return nil, fmt.Errorf("GPIO controller is required")
	}

	// Set defaults
	if cfg.MonitorInterval == 0 {
		cfg.MonitorInterval = defaultMonitorInterval
	}

	m := &Monitor{
		gpio:           cfg.GPIO,
		fanPin:         cfg.FanControlPin,
		throttlePin:    cfg.ThrottlePin,
		cpuTemp:        cfg.CPUTempPath,
		gpuTemp:        cfg.GPUTempPath,
		ambientTemp:    cfg.AmbientTempPath,
		monitorInterval: cfg.MonitorInterval,
		onWarning:      cfg.OnWarning,
		onCritical:     cfg.OnCritical,
	}

	// Configure GPIO pins
	if m.fanPin != "" {
		pin := gpio.PinIO(&gpio.BasicPin{})
		if err := m.gpio.ConfigurePin(m.fanPin, pin, gpio.Float); err != nil {
			return nil, fmt.Errorf("failed to configure fan pin: %w", err)
		}
	}
	if m.throttlePin != "" {
		pin := gpio.PinIO(&gpio.BasicPin{})
		if err := m.gpio.ConfigurePin(m.throttlePin, pin, gpio.Float); err != nil {
			return nil, fmt.Errorf("failed to configure throttle pin: %w", err)
		}
	}

	return m, nil
}

// GetState returns the current thermal state
func (m *Monitor) GetState() ThermalState {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state
}

// Monitor starts monitoring thermal state in the background
func (m *Monitor) Monitor(ctx context.Context) error {
	ticker := time.NewTicker(m.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.updateThermalState(); err != nil {
				return fmt.Errorf("failed to update thermal state: %w", err)
			}
		}
	}
}