package secure

import (
	"context"
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

// TamperState represents the current tamper detection status
type TamperState struct {
	CaseOpen      bool
	MotionDetected bool
	VoltageNormal  bool
	LastCheck      time.Time
}

// Config holds the configuration for the security manager
type Config struct {
	GPIO          *gpio.Controller
	CaseSensor    string
	MotionSensor  string
	VoltageSensor string
	DeviceID      string
	StateStore    StateStore
	OnTamper      func(TamperState)
}

// StateStore defines the interface for persisting security state
type StateStore interface {
	// SaveState persists the current security state
	SaveState(ctx context.Context, deviceID string, state TamperState) error

	// LoadState retrieves the last known security state
	LoadState(ctx context.Context, deviceID string) (TamperState, error)

	// LogEvent records a security event
	LogEvent(ctx context.Context, deviceID string, eventType string, details interface{}) error
}

// Event represents a security event
type Event struct {
	DeviceID  string
	Type      string
	Timestamp time.Time
	Details   interface{}
}