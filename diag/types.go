package diag

import (
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
	"github.com/wrale/wrale-fleet-metal-hw/power"
	"github.com/wrale/wrale-fleet-metal-hw/secure"
	"github.com/wrale/wrale-fleet-metal-hw/thermal"
)

// TestType identifies different hardware diagnostic tests
type TestType string

const (
	TestGPIO     TestType = "GPIO"
	TestPower    TestType = "POWER"
	TestThermal  TestType = "THERMAL"
	TestSecurity TestType = "SECURITY"
)

// TestResult represents a hardware test outcome
type TestResult struct {
	Type        TestType
	Component   string
	Status      TestStatus
	Reading     float64
	Expected    float64
	Description string
	Error       error
	Timestamp   time.Time
}

// TestStatus represents the outcome of a hardware test
type TestStatus string

const (
	StatusPass    TestStatus = "PASS"
	StatusFail    TestStatus = "FAIL"
	StatusWarning TestStatus = "WARNING"
	StatusSkipped TestStatus = "SKIPPED"
)

// Config holds the hardware diagnostics configuration
type Config struct {
	// Hardware subsystem interfaces
	GPIO     *gpio.Controller
	Power    *power.Manager
	Thermal  *thermal.Monitor
	Security *secure.Manager

	// Test parameters
	GPIOPins     []string       // GPIO pins to test
	LoadTestTime time.Duration  // Duration for power load tests
	MinVoltage   float64        // Minimum acceptable voltage
	TempRange    [2]float64     // Valid temperature range
	Retries      int           // Number of test retries

	// Optional callbacks
	OnTestComplete func(TestResult)
}

// RawReadings holds direct sensor readings
type RawReadings struct {
	GPIOStates     map[string]bool // Raw GPIO pin states
	Voltages       []float64       // Raw voltage measurements
	Temperatures   []float64       // Raw temperature readings
	SecurityInputs []bool         // Raw security sensor states
}