# Wrale Fleet Metal Hardware

Hardware management layer for Wrale Fleet. Handles GPIO, power management, physical security, and thermal management for Raspberry Pi devices.

## Scope

### In Scope
- Direct hardware interactions and abstractions
- Hardware state management and monitoring
- Hardware-level safety controls
- Physical security monitoring
- Raw sensor data collection
- Hardware simulation for testing

### Out of Scope
- System-level orchestration (fleet-metal-core)
- Fleet-wide state management
- Data persistence and synchronization
- Network communication
- Business logic and policy decisions
- User interfaces and APIs
- Advanced analytics and diagnostics (fleet-metal-diag)

## Hardware Subsystems

### GPIO Management
- Pin control and monitoring
- Interrupt handling
- PWM support

### Power Management
- Multiple power sources
- Battery monitoring
- Voltage monitoring
- Power state tracking

### Thermal Management
- CPU/GPU temperature monitoring
- Fan control
- Thermal throttling

### Physical Security
- Case intrusion detection
- Motion detection
- Tamper monitoring

## Hardware Testing

See testing guides for:
- Hardware simulation setup
- RPi 3B+ validation
- RPi 5 deployment

## License

Apache License 2.0