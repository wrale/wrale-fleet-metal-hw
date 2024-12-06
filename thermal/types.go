package thermal

import (
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

// Temperature thresholds in Celsius
const (
	// Warning thresholds
	cpuTempWarning  = 70.0
	gpuTempWarning  = 70.0
	ambientWarning  = 45.0

	// Critical thresholds
	cpuTempCritical  = 80.0
	gpuTempCritical  = 80.0
	ambientCritical  = 50.0

	// Default monitoring interval
	defaultMonitorInterval = 1 * time.Second
)

// ThermalState represents current thermal conditions
type ThermalState struct {
	CPUTemp     float64   // CPU temperature in Celsius
	GPUTemp     float64   // GPU temperature in Celsius
	AmbientTemp float64   // Ambient temperature in Celsius
	FanSpeed    int       // Current fan speed percentage
	Throttled   bool      // Whether system is throttled
	Warnings    []string  // Active thermal warnings
	UpdatedAt   time.Time // Last update timestamp
}

// Config holds thermal monitor configuration
type Config struct {
	GPIO             *gpio.Controller
	MonitorInterval  time.Duration
	CPUTempPath      string  // sysfs path to CPU temperature
	GPUTempPath      string  // sysfs path to GPU temperature
	AmbientTempPath  string  // sysfs path to ambient temperature sensor
	FanControlPin    string  // GPIO pin for fan control
	ThrottlePin      string  // GPIO pin for throttling control
	OnWarning        func(ThermalState)  // Callback for warning conditions
	OnCritical       func(ThermalState)  // Callback for critical conditions
}