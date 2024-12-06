package thermal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

// Monitor handles thermal monitoring and control
type Monitor struct {
	mux   sync.RWMutex
	state ThermalState

	// Hardware interface
	gpio       *gpio.Controller
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

	if m.fanPin != "" {
		if err := m.InitializeFanControl(); err != nil {
			return nil, fmt.Errorf("failed to initialize fan: %w", err)
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

// SetFanSpeed sets the fan speed to a specific percentage
func (m *Monitor) SetFanSpeed(speed int) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	
	m.setFanSpeedLocked(speed)
	return nil
}