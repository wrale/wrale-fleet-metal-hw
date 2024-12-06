# Hardware Simulation Guide

## Overview

The hardware simulation layer enables development and testing without physical Raspberry Pi hardware. It provides:
- GPIO simulation
- Power state simulation
- Thermal behavior simulation
- Security sensor simulation
- Hardware fault injection

## Architecture

```
Application Code
      ↓
Hardware Interface
      ↓
Simulation Layer ↔ Real Hardware
```

## Quick Start

```bash
# Start simulation environment
make sim-start

# Run tests with simulation
go test ./... -tags=simulation

# Simulate hardware events
go run cmd/hwsim/main.go trigger --event power_loss

# Stop simulation
make sim-stop
```

## Simulated Components

### GPIO Simulation
```
/tmp/wrale-sim/gpio/
  ├── pin_{N}/
  │   ├── value
  │   ├── direction
  │   ├── edge
  │   └── active_low
  └── pwm_{N}/
      ├── period
      ├── duty_cycle
      └── enable
```

### Power Simulation
```
/tmp/wrale-sim/power/
  ├── voltage
  ├── current
  ├── battery_level
  └── sources/
      ├── main
      ├── battery
      └── solar
```

### Thermal Simulation
```
/tmp/wrale-sim/thermal/
  ├── cpu_temp
  ├── gpu_temp
  ├── ambient_temp
  └── fan_speed
```

### Security Simulation
```
/tmp/wrale-sim/security/
  ├── case_sensor
  ├── motion_sensor
  └── voltage_monitor
```

## Usage Examples

### GPIO Testing
```go
func TestGPIO(t *testing.T) {
    // Create GPIO controller with simulation
    ctrl := NewGPIOController(WithSimulation())
    
    // Test pin operations
    err := ctrl.SetPinState("test_pin", true)
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify state
    state, err := ctrl.GetPinState("test_pin")
    if err != nil || !state {
        t.Error("Pin state mismatch")
    }
}
```

### Power Testing
```go
func TestPowerFailure(t *testing.T) {
    // Setup power manager with simulation
    mgr := NewPowerManager(WithSimulation())
    
    // Simulate power loss
    sim := GetSimulation()
    sim.TriggerPowerLoss()
    
    // Verify response
    state := mgr.GetState()
    if state.CurrentSource != BatteryPower {
        t.Error("Failed to switch to battery")
    }
}
```

### Thermal Testing
```go
func TestThermalThrottling(t *testing.T) {
    // Create thermal monitor with simulation
    mon := NewThermalMonitor(WithSimulation())
    
    // Simulate temperature rise
    sim := GetSimulation()
    sim.SetCPUTemp(85.0)
    
    // Verify response
    state := mon.GetState()
    if !state.Throttled {
        t.Error("Throttling not enabled")
    }
}
```

## Fault Injection

### Available Faults
1. Power System
   - Power loss
   - Voltage sag
   - Battery failure
   - Current spike

2. Thermal System
   - Temperature spike
   - Fan failure
   - Sensor malfunction
   - Throttling failure

3. Security System
   - Case intrusion
   - Motion detection
   - Voltage tampering
   - Sensor failure

### Example: Fault Injection
```go
func TestFaultHandling(t *testing.T) {
    sim := GetSimulation()
    
    // Inject fault
    sim.InjectFault(FaultDef{
        Type: PowerFault,
        Mode: VoltageSag,
        Duration: 100 * time.Millisecond,
        Magnitude: 0.8, // 80% voltage
    })
    
    // Verify handling
    // ...
}
```

## Performance Simulation

### Simulated Metrics
- GPIO timing
- Power consumption
- Temperature curves
- Sensor latency

### Example: Performance Testing
```go
func TestPWMPerformance(t *testing.T) {
    sim := GetSimulation()
    
    // Configure performance characteristics
    sim.SetTiming(PWMTiming{
        Resolution: 100 * time.Nanosecond,
        Jitter: 50 * time.Nanosecond,
    })
    
    // Run performance test
    // ...
}
```

## Limitations

### Known Limitations
1. Timing accuracy
   - No real-time guarantees
   - Simplified interrupt handling
   - PWM approximation

2. Hardware Specifics
   - Limited voltage rail simulation
   - Simplified thermal model
   - Basic security simulation

3. Environmental Factors
   - No EMI simulation
   - Simple thermal transfer
   - Basic power dynamics

### Best Practices
1. Use simulation for:
   - Basic functionality testing
   - Fault handling
   - Initial development
   - CI/CD pipelines

2. Always validate on real hardware:
   - Timing-critical features
   - Power management
   - Thermal behavior
   - Security features