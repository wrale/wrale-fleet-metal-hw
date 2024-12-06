package secure

import (
	"context"
	"fmt"
	"sync"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

// Manager handles physical security monitoring and response
type Manager struct {
	mux sync.RWMutex
	gpio *gpio.Controller

	// Sensor pin names
	caseSensor   string
	motionSensor string
	voltSensor   string

	// Device identification
	deviceID string

	// State management
	state      TamperState
	stateStore StateStore

	// Callbacks for security events
	onTamper func(TamperState)
}

// New creates a new security manager
func New(cfg Config) (*Manager, error) {
	if cfg.GPIO == nil {
		return nil, fmt.Errorf("GPIO controller is required")
	}
	if cfg.DeviceID == "" {
		return nil, fmt.Errorf("device ID is required")
	}

	m := &Manager{
		gpio:         cfg.GPIO,
		caseSensor:   cfg.CaseSensor,
		motionSensor: cfg.MotionSensor,
		voltSensor:   cfg.VoltageSensor,
		deviceID:     cfg.DeviceID,
		stateStore:   cfg.StateStore,
		onTamper:     cfg.OnTamper,
	}

	// Load last known state if store is available
	if m.stateStore != nil {
		if state, err := m.stateStore.LoadState(context.Background(), m.deviceID); err == nil {
			m.state = state
		}
	}

	return m, nil
}

// GetState returns the current security state
func (m *Manager) GetState() TamperState {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state
}