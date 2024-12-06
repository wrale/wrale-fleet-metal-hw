package power

import (
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

// PowerSource represents a power supply type
type PowerSource string

const (
	MainPower    PowerSource = "MAIN"
	BatteryPower PowerSource = "BATTERY"
	SolarPower   PowerSource = "SOLAR"

	// Default monitoring interval
	defaultMonitorInterval   = 1 * time.Second
	defaultStabilityInterval = 100 * time.Millisecond
)

// PowerState represents the current power status
type PowerState struct {
	BatteryLevel     float64
	Charging         bool
	Voltage          float64
	CurrentSource    PowerSource
	AvailablePower   map[PowerSource]bool
	PowerConsumption float64 // in watts
	UpdatedAt        time.Time

	// Enhanced stability metrics
	StabilityMetrics *StabilityMetrics `json:",omitempty"`
}

// StabilityMetrics provides detailed power quality measurements
type StabilityMetrics struct {
	VoltageRipple     float64       // Peak-to-peak voltage variation
	MinVoltage        float64       // Minimum voltage in sample window
	MaxVoltage        float64       // Maximum voltage in sample window
	AverageVoltage    float64       // Moving average voltage
	PowerCycles       int           // Number of power cycles detected
	LastCycleDuration time.Duration // Duration of last power cycle
	CurrentSpikes     int           // Number of current spikes detected
	MaxCurrentSpike   float64       // Maximum current spike amplitude
	Warnings          []string      // Active power quality warnings
	UpdatedAt         time.Time     // Last metrics update
}

// StabilityConfig holds power stability monitoring configuration
type StabilityConfig struct {
	// Minimum samples to keep for stability analysis
	SampleWindow int
	// How frequently to sample power metrics
	SampleInterval time.Duration
	// Voltage ripple threshold (volts)
	RippleThreshold float64
	// Maximum allowed current draw (amps)
	CurrentThreshold float64
	// Callback for stability events
	OnStabilityEvent func(StabilityEvent)
}

// StabilityEvent represents a power quality incident
type StabilityEvent struct {
	Timestamp time.Time
	Type      StabilityEventType
	Reading   float64
	Threshold float64
	Source    PowerSource
	Details   string
}

// StabilityEventType identifies different stability events
type StabilityEventType string

const (
	EventVoltageRipple  StabilityEventType = "VOLTAGE_RIPPLE"
	EventVoltageSag     StabilityEventType = "VOLTAGE_SAG"
	EventCurrentSpike   StabilityEventType = "CURRENT_SPIKE"
	EventPowerCycle     StabilityEventType = "POWER_CYCLE"
	EventSourceFailover StabilityEventType = "SOURCE_FAILOVER"
)

// Config holds power manager configuration
type Config struct {
	GPIO            *gpio.Controller
	MonitorInterval time.Duration
	PowerPins       map[PowerSource]string // GPIO pins for power sources
	BatteryADCPath  string                 // sysfs path to battery ADC
	VoltageADCPath  string                 // sysfs path to voltage ADC
	CurrentADCPath  string                 // sysfs path to current sensor ADC
	OnPowerCritical func(PowerState)       // Callback for critical power events

	// Stability monitoring configuration
	StabilityConfig *StabilityConfig
}
