# Wrale Fleet Metal Hardware

[![Go](https://github.com/wrale/wrale-fleet-metal-hw/actions/workflows/go.yml/badge.svg)](https://github.com/wrale/wrale-fleet-metal-hw/actions/workflows/go.yml)
[![Lint](https://github.com/wrale/wrale-fleet-metal-hw/actions/workflows/lint.yml/badge.svg)](https://github.com/wrale/wrale-fleet-metal-hw/actions/workflows/lint.yml)

Pure hardware management layer for Wrale Fleet. Handles direct hardware interactions, raw sensor data, and hardware-level safety for Raspberry Pi devices. Part of the Wrale Fleet Metal project.

## Feature Status

| Feature | Developed | Unit Written | Unit Passing | HW Sim | HW Tested | Required Hardware |
|---------|-----------|--------------|--------------|--------|-----------|------------------|
| **GPIO Management** |
| - Pin Control | ✅ | ✅ | ✅ | ✅ | ❓ | None |
| - Interrupt Handling | ✅ | ✅ | ✅ | ✅ | ❓ | None |
| - PWM Support | ✅ | ✅ | ✅ | ✅ | ❓ | None |
| - Pull-up/down Config | ✅ | ✅ | ✅ | ✅ | ❓ | None |
| **Power Management** |
| - Multiple Sources | ✅ | ✅ | ✅ | ✅ | ❓ | Power Distribution Board |
| - Battery Monitoring | ✅ | ✅ | ✅ | ❓ | ❓ | Battery HAT / UPS |
| - Voltage/Current | ✅ | ✅ | ✅ | ❓ | ❓ | INA219 or Similar ADC |
| - Load Testing | ✅ | ✅ | ✅ | ❓ | ❓ | Load Banks, Power Meter |
| - Power Stability | ✅ | ✅ | ✅ | ❓ | ❓ | Voltage Monitor Board |
| **Thermal Management** |
| - Temperature Reading | ✅ | ✅ | ✅ | ✅ | ❓ | None (Built-in) |
| - PWM Fan Control | ✅ | ✅ | ✅ | ✅ | ❓ | PWM-compatible Fan |
| - Thermal Throttling | ✅ | ✅ | ✅ | ✅ | ❓ | None (Built-in) |
| - Multi-zone Monitoring | ✅ | ✅ | ✅ | ✅ | ❓ | External Temp Sensors |
| **Physical Security** |
| - Case Intrusion | ✅ | ✅ | ✅ | ✅ | ❓ | Case Switch Sensor |
| - Motion Detection | ✅ | ✅ | ✅ | ✅ | ❓ | PIR Motion Sensor |
| - Voltage Monitoring | ✅ | ✅ | ✅ | ✅ | ❓ | Voltage Monitor Board |
| - Tamper Response | ✅ | ✅ | ✅ | ✅ | ❓ | Security GPIO Board |
| **Hardware Diagnostics** |
| - Component Testing | ✅ | ✅ | ✅ | ✅ | ❓ | Varies by Component |
| - Sensor Validation | ✅ | ✅ | ✅ | ✅ | ❓ | Calibration Tools |
| - Load Testing | ✅ | ✅ | ✅ | ❓ | ❓ | Load Banks, Power Meter |
| - Fault Detection | ✅ | ✅ | ✅ | ❓ | ❓ | Sensor Array |

Legend:
- ✅ Completed/Verified
- ❓ To Be Verified
- ❌ Not Started

## Scope

### In Scope
- Direct hardware interactions and abstractions
- Raw sensor data collection and validation
- Hardware-level safety controls
- Physical security monitoring
- Hardware simulation for testing
- Basic hardware state management

### Out of Scope
- System-level orchestration (fleet-metal-core)
- Fleet-wide state management
- Data persistence and synchronization
- Network communication
- Business logic and policy decisions
- User interfaces and APIs
- Advanced analytics (fleet-metal-diag)
- Environmental control logic

## Hardware Subsystems

### GPIO Management
- Raw pin control and monitoring
- Hardware interrupt handling
- PWM support with frequency control
- Pull-up/down configuration

### Power Management
- Multiple power source management
- Battery level monitoring
- Voltage and current monitoring 
- Power stability monitoring
- Load testing capabilities
- Hardware-level power safety

### Thermal Management
- CPU/GPU temperature monitoring
- PWM-based fan speed control
- Hardware thermal throttling
- Raw temperature data collection

### Physical Security
- Case intrusion detection
- Motion sensor monitoring
- Voltage tamper detection
- Raw security sensor data

### Hardware Diagnostics
- Hardware validation testing
- Raw sensor verification
- Hardware simulation support
- Basic hardware health checks
- Power load testing

## Testing Features

### Simulation Mode
The package includes a simulation mode that allows testing without physical hardware:
- GPIO simulation with pin state and pull configuration
- PWM duty cycle and fan control simulation
- Basic power state simulation
- Thermal state tracking
- Security sensor simulation

### Hardware Testing
- Full simulation mode for development and testing
- Unit tests for all subsystems
- Integration test coverage
- Physical hardware validation (requires RPi)

### Continuous Integration
The project includes GitHub Actions workflows for:
- Automated testing on each PR and push
- Go linting and static analysis
- Race condition detection
- Code quality checks

## Directory Structure
```
.
├── gpio/       # GPIO and PWM control
├── power/      # Power management
├── secure/     # Physical security
├── thermal/    # Temperature control
└── diag/       # Hardware diagnostics
```

## Hardware Requirements

### Raspberry Pi Support
- Full support: RPi 3B+ (WIP)
- Testing: RPi 5 (WIP)
- Memory: 512MB minimum
- Storage: Basic system only

### Power Requirements
- Input: 5V DC
- Current: 2A minimum
- Battery backup support
- Multiple power source support

### Environmental
- Operating temp: -10°C to 50°C
- Humidity: Up to 80% non-condensing
- Enclosure: IP65 or better recommended

### Additional Hardware (Based on Features)
- PWM-compatible cooling fan
- Battery HAT or UPS system
- Power monitoring HAT/board (INA219 or similar)
- Temperature sensor array
- Security sensors (case, motion, voltage)
- Load banks for power testing
- Calibration tools for sensors

## Integration

This package provides the hardware abstraction layer for the Wrale Fleet Metal system. It should be consumed by fleet-metal-core for system-level management. Direct hardware access should only occur through this package.

## Development

### Prerequisites
- Go 1.21+
- Access to RPi hardware or simulation environment
- Basic electronics knowledge
- Required hardware components based on features needed

### Testing
```bash
# Run all tests in simulation mode
go test -v -race ./...

# Run linting checks
golangci-lint run

# Test specific hardware subsystem
go test -v -race ./power/...
```

See [Hardware Testing Guide](docs/HARDWARE_TESTING.md) for physical device testing details.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

Apache License 2.0
