package power

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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

	// Internal state
	shutdownInitiated bool
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
		if err := m.gpio.ConfigurePin(pin, nil); err != nil {
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