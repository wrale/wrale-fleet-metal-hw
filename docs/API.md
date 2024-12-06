# Wrale Fleet Metal Hardware API Guide

This document describes the public interfaces and usage patterns for the Wrale Fleet Metal Hardware library.

## Core Concepts

### Hardware Subsystems
The library provides five core subsystems:
- GPIO Management (gpio)
- Power Management (power)
- Thermal Management (thermal)
- Physical Security (secure) 
- Hardware Diagnostics (diag)

Each subsystem follows consistent patterns:
1. A main controller/manager type
2. A configuration struct for initialization
3. State monitoring with regular updates
4. Event callbacks for critical conditions
5. Simulation mode support for testing

## Initialization Pattern

All subsystems follow a similar initialization pattern:

```go
// 1. Create GPIO controller first (required by other subsystems)
gpioCtrl, err := gpio.New(
    gpio.WithSimulation(), // Optional: enable simulation mode
)
if err != nil {
    return err
}

// 2. Initialize required subsystem
manager, err := subsystem.New(subsystem.Config{
    GPIO: gpioCtrl,
    // ... other config options
})
if err != nil {
    return err
}

// 3. Start monitoring (if applicable)
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go manager.Monitor(ctx)
```

## GPIO Management

The GPIO subsystem provides low-level hardware access.

### Key Types
```go
// Configure a GPIO pin
err := ctrl.ConfigurePin("pin_name", pin, gpio.Pull)

// Configure PWM
err := ctrl.ConfigurePWM("fan", pin, gpio.PWMConfig{
    Frequency: 25000,  // Hz
    DutyCycle: 50,    // 0-100
    Pull: gpio.PullNone,
})

// Control PWM
err := ctrl.SetPWMDutyCycle("fan", 75)  // 0-100
```

### Simulation Support
```go
ctrl, err := gpio.New(gpio.WithSimulation())
```

## Power Management

Handles power sources, monitoring, and safety.

### Key Types
```go
type PowerSource string

const (
    MainPower    PowerSource = "MAIN"
    BatteryPower PowerSource = "BATTERY"
    SolarPower   PowerSource = "SOLAR"
)

type PowerState struct {
    BatteryLevel     float64
    Charging         bool
    Voltage          float64
    CurrentSource    PowerSource
    AvailablePower   map[PowerSource]bool
    PowerConsumption float64
    UpdatedAt        time.Time
    StabilityMetrics *StabilityMetrics
}
```

### Usage
```go
mgr, err := power.New(power.Config{
    GPIO:            gpioCtrl,
    MonitorInterval: time.Second,
    PowerPins: map[PowerSource]string{
        MainPower:    "main_power",
        BatteryPower: "battery_power",
    },
    OnPowerCritical: func(state PowerState) {
        // Handle critical power events
    },
})

// Get current state
state := mgr.GetState()
```

## Thermal Management

Monitors temperatures and controls cooling.

### Key Types
```go
type ThermalState struct {
    CPUTemp     float64   // Celsius
    GPUTemp     float64   // Celsius
    AmbientTemp float64   // Celsius
    FanSpeed    uint32    // Percentage (0-100)
    Throttled   bool
    Warnings    []string
    UpdatedAt   time.Time
}
```

### Usage
```go
monitor, err := thermal.New(thermal.Config{
    GPIO:            gpioCtrl,
    MonitorInterval: time.Second,
    FanControlPin:   "fan_pwm",
    ThrottlePin:     "throttle",
    OnWarning: func(state ThermalState) {
        // Handle thermal warnings
    },
    OnCritical: func(state ThermalState) {
        // Handle critical temperature
    },
})

// Control fan manually (if needed)
err := monitor.SetFanSpeed(75)  // 0-100
```

## Physical Security

Handles tamper detection and security monitoring.

### Key Types
```go
type TamperState struct {
    CaseOpen       bool
    MotionDetected bool
    VoltageNormal  bool
    LastCheck      time.Time
}

// Interface for security state persistence
type StateStore interface {
    SaveState(ctx context.Context, deviceID string, state TamperState) error
    LoadState(ctx context.Context, deviceID string) (TamperState, error)
    LogEvent(ctx context.Context, deviceID string, eventType string, details interface{}) error
}
```

### Usage
```go
mgr, err := secure.New(secure.Config{
    GPIO:          gpioCtrl,
    CaseSensor:    "case_switch",
    MotionSensor:  "pir_sensor",
    VoltageSensor: "voltage_monitor",
    DeviceID:      "device_1",
    StateStore:    myStateStore,
    OnTamper: func(state TamperState) {
        // Handle tamper events
    },
})
```

## Hardware Diagnostics

Provides comprehensive hardware testing capabilities.

### Key Types
```go
type TestType string
const (
    TestGPIO     TestType = "GPIO"
    TestPower    TestType = "POWER"
    TestThermal  TestType = "THERMAL"
    TestSecurity TestType = "SECURITY"
)

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
```

### Usage
```go
diag, err := diag.New(diag.Config{
    GPIO:     gpioCtrl,
    Power:    powerMgr,
    Thermal:  thermalMon,
    Security: secureMgr,
    OnTestComplete: func(result TestResult) {
        // Handle test results
    },
})

// Run specific tests
err := diag.TestPower(ctx)
err := diag.TestThermal(ctx)

// Run full diagnostic suite
err := diag.RunAll(ctx)
```

## Error Handling

All subsystems follow consistent error handling patterns:
1. Configuration errors during initialization
2. Hardware access errors during operation
3. State transition errors
4. Resource cleanup errors

Example:
```go
if err := mgr.SetFanSpeed(speed); err != nil {
    switch {
    case errors.Is(err, gpio.ErrDisabled):
        // Handle disabled state
    case errors.Is(err, gpio.ErrInvalidValue):
        // Handle invalid input
    default:
        // Handle unexpected error
    }
}
```

## Resource Cleanup

Always close resources when done:
```go
defer gpioCtrl.Close()
defer powerMgr.Close()
defer thermalMon.Close()
defer secureMgr.Close()
```

## Threading Considerations

1. All public methods are thread-safe
2. State updates happen atomically
3. Callbacks are executed in separate goroutines
4. Monitor loops should be managed with context cancellation

## Development Patterns

### Testing Without Hardware
```go
ctrl, err := gpio.New(gpio.WithSimulation())
if err != nil {
    return err
}

// Now other subsystems will work in simulation mode
mgr, err := power.New(power.Config{
    GPIO: ctrl,
    // ...
})
```

### Handling Critical Events

Implement the appropriate callbacks for your use case:
- OnPowerCritical for power events
- OnWarning/OnCritical for thermal events
- OnTamper for security events
- OnTestComplete for diagnostics

### State Monitoring

Poll states periodically or use callbacks for changes:
```go
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-ticker.C:
        state := mgr.GetState()
        // Process state
    }
}
```

## Best Practices

1. **Initialization Order**
   - Create GPIO controller first
   - Initialize other subsystems in dependency order
   - Start monitoring routines last

2. **Error Handling**
   - Check initialization errors
   - Monitor hardware errors
   - Handle cleanup errors
   - Log or report all errors

3. **Resource Management**
   - Use context for cancellation
   - Properly close all resources
   - Clean up in reverse initialization order

4. **State Management**
   - Poll states at appropriate intervals
   - Don't poll too frequently
   - Use callbacks for critical events

5. **Testing**
   - Use simulation mode for initial testing
   - Test error conditions
   - Verify state transitions
   - Test cleanup handling

## Security Notes

1. Physical security events should be handled promptly
2. Validate all inputs to public methods
3. Implement appropriate access controls
4. Log security-relevant events
5. Handle tamper events appropriately

## Performance Considerations

1. Default monitoring intervals are reasonable for most uses
2. Adjust intervals based on your needs
3. Be mindful of callback execution time
4. Consider buffering rapid state changes
5. Use appropriate timeouts for operations

