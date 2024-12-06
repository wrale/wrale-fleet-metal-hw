package power

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

// Manager handles power-related operations
type Manager struct {
	mux   sync.RWMutex
	state PowerState

	// Hardware interface
	gpio      *gpio.Controller
	powerPins map[PowerSource]string

	// ADC paths
	batteryADC string
	voltageADC string
	currentADC string

	// Configuration
	monitorInterval time.Duration
	onPowerCritical func(PowerState)
}

// New creates a new power manager
func New(cfg Config) (*Manager, error) {
	if cfg.GPIO == nil {
		return nil, fmt.Errorf("GPIO controller is required")
	}

	// Set defaults
	if cfg.MonitorInterval == 0 {
		cfg.MonitorInterval = defaultMonitorInterval
	}

	m := &Manager{
		gpio:            cfg.GPIO,
		powerPins:       cfg.PowerPins,
		batteryADC:      cfg.BatteryADCPath,
		voltageADC:      cfg.VoltageADCPath,
		currentADC:      cfg.CurrentADCPath,
		monitorInterval: cfg.MonitorInterval,
		onPowerCritical: cfg.OnPowerCritical,
		state: PowerState{
			AvailablePower: make(map[PowerSource]bool),
		},
	}

	// Initialize power source pins
	for source, pin := range cfg.PowerPins {
		if err := m.gpio.ConfigurePin(pin, nil, gpio.PullNone); err != nil {
			return nil, fmt.Errorf("failed to configure power pin %s: %w", pin, err)
		}
		m.state.AvailablePower[source] = false
	}

	return m, nil
}

// GetState returns the current power state
func (m *Manager) GetState() PowerState {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state
}

// Monitor starts monitoring power state in the background
func (m *Manager) Monitor(ctx context.Context) error {
	ticker := time.NewTicker(m.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.updatePowerState(ctx); err != nil {
				return fmt.Errorf("failed to update power state: %w", err)
			}
		}
	}
}

// updatePowerState reads current power status
func (m *Manager) updatePowerState(ctx context.Context) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	// Check power sources
	for source, pin := range m.powerPins {
		available, err := m.gpio.GetPinState(pin)
		if err != nil {
			return fmt.Errorf("failed to check power source %s: %w", source, err)
		}
		m.state.AvailablePower[source] = available
	}

	m.state.UpdatedAt = time.Now()
	return nil
}
