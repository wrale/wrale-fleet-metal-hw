package power

import (
	"time"
)

// PowerSource represents a power supply type
type PowerSource string

const (
	MainPower    PowerSource = "MAIN"
	BatteryPower PowerSource = "BATTERY"
	SolarPower   PowerSource = "SOLAR"
	
	// Default monitoring interval
	defaultMonitorInterval = 1 * time.Second

	// Critical power thresholds
	criticalBatteryLevel = 10.0  // 10% battery remaining
	criticalVoltage     = 4.8   // 4.8V (assuming 5V system)
)

// PowerState represents the current power status
type PowerState struct {
	BatteryLevel    float64
	Charging        bool
	Voltage         float64
	CurrentSource   PowerSource
	AvailablePower  map[PowerSource]bool
	PowerConsumption float64 // in watts
	UpdatedAt       time.Time
}

// Config holds power manager configuration
type Config struct {
	GPIO            *gpio.Controller
	MonitorInterval time.Duration
	PowerPins       map[PowerSource]string // GPIO pins for power sources
	BatteryADCPath  string  // sysfs path to battery ADC
	VoltageADCPath  string  // sysfs path to voltage ADC
	CurrentADCPath  string  // sysfs path to current sensor ADC
	OnPowerCritical func(PowerState) // Callback for critical power events
}