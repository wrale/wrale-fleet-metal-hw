# Wrale Fleet Metal Hardware

Pure hardware management layer for Wrale Fleet. Handles direct hardware interactions, raw sensor data, and hardware-level safety for Raspberry Pi devices. Part of the Wrale Fleet Metal project.

## Feature Status

| Feature | Developed | Unit Written | Unit Passing | HW Sim | HW Tested |
|---------|-----------|--------------|--------------|--------|------------|
| **GPIO Management** |
| - Pin Control | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Interrupt Handling |  ✅ | ✅ | ❓ | ❓ | ❓ |
| - PWM Support | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Pull-up/down Config | ✅ | ✅ | ❓ | ❓ | ❓ |
| **Power Management** |
| - Multiple Sources | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Battery Monitoring | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Voltage/Current | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Load Testing | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Power Stability | ✅ | ✅ | ❓ | ❓ | ❓ |
| **Thermal Management** |
| - Temperature Reading | ✅ | ✅ | ❓ | ❓ | ❓ |
| - PWM Fan Control | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Thermal Throttling | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Multi-zone Monitoring | ✅ | ✅ | ❓ | ❓ | ❓ |
| **Physical Security** |
| - Case Intrusion | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Motion Detection | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Voltage Monitoring | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Tamper Response | ✅ | ✅ | ❓ | ❓ | ❓ |
| **Hardware Diagnostics** |
| - Component Testing | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Sensor Validation | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Load Testing | ✅ | ✅ | ❓ | ❓ | ❓ |
| - Fault Detection | ✅ | ✅ | ❓ | ❓ | ❓ |

Legend:
- ✅ Completed
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

## Integration

This package provides the hardware abstraction layer for the Wrale Fleet Metal system. It should be consumed by fleet-metal-core for system-level management. Direct hardware access should only occur through this package.

## Development

### Prerequisites
- Go 1.21+
- Access to RPi hardware or simulation environment
- Basic electronics knowledge

### Testing
```bash
# Run all hardware tests with simulation
go test ./...

# Test specific hardware subsystem
go test ./power/...
```

See [Hardware Testing Guide](docs/HARDWARE_TESTING.md) for physical device testing.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

Apache License 2.0
