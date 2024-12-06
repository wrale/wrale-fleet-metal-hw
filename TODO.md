# TODO - Wrale Fleet Metal Hardware

## Documentation Priority
- [x] Create basic simulation environment
- [ ] Create detailed hardware testing guide
  - Testing with real RPi hardware
  - Hardware validation procedures
  - Common issues and solutions
- [ ] Write RPi 5 validation guide
  - Hardware differences from 3B+
  - Testing procedures
  - Known limitations
  - Performance considerations
- [ ] Complete simulation documentation
  - Available simulated features:
    - GPIO pin state and pull configuration
    - PWM duty cycle and fan control
    - Basic power state management
    - Thermal state tracking
  - Simulation limitations:
    - No real-time PWM generation
    - Simplified thermal modeling
    - Basic power state tracking
  - Adding new simulated components
  - Testing practices with simulation

## Testing Gaps
- [x] Basic unit test coverage with simulation
- [ ] Full test coverage for power load testing
  - Add simulated load response
  - Voltage sag under load
  - Battery failover behavior
- [ ] Integration tests between subsystems
  - Power/Thermal interactions
  - Security/Power relationships
  - GPIO state persistence
- [ ] Edge case fault injection tests
  - Simulated hardware failures
  - Data corruption scenarios 
  - Race condition testing
- [ ] Physical deployment test suite
  - Hardware burn-in tests
  - Long-term stability monitoring
  - Environmental stress testing

## Hardware Support
- [ ] Complete RPi 5 hardware validation
- [ ] Test with different power supplies
  - Official RPi PSU
  - High-amp USB-C supplies
  - Battery backup systems
- [ ] Validate different fan types
  - PWM response curves
  - Speed vs airflow
  - Acoustic profiles
- [ ] Test various security sensors
  - Case intrusion switches
  - Motion detectors
  - Voltage monitors

## Performance
- [ ] PWM frequency calibration
  - Software vs hardware PWM timing
  - Jitter measurement
  - Duty cycle accuracy
- [ ] Power measurement calibration
  - Voltage measurement accuracy
  - Current sensor calibration
  - Power calculation validation
- [ ] Temperature sensor calibration
  - Sensor accuracy verification
  - Cross-sensor correlation
  - Response time measurement
- [ ] Interrupt latency testing
  - GPIO interrupt timing
  - Event propagation delays
  - System load impact

## Safety Enhancements
- [ ] Add thermal runaway protection
  - Temperature trend analysis
  - Emergency shutdown thresholds
  - Fail-safe cooling modes
- [ ] Enhance power failover logic
  - Clean power source switching
  - Battery life optimization
  - Graceful shutdown timing
- [ ] Improve voltage sag detection
  - Fast sampling for transients
  - Sag depth measurement
  - Recovery monitoring
- [ ] Add current spike protection
  - Inrush current limiting
  - Over-current response
  - Load shedding logic

## Future Considerations
- [ ] Support for additional RPi models
  - CM4 support
  - Zero series support
  - Future model compatibility
- [ ] Alternative sensor support
  - I2C temperature sensors
  - SPI power monitors
  - External security devices
- [ ] Enhanced PWM capabilities
  - Variable frequency support
  - Multiple channel sync
  - Complex waveforms
- [ ] Power efficiency improvements
  - Dynamic frequency scaling
  - Load-based optimization
  - Sleep mode support

## Integration Support
- [ ] Document integration patterns for fleet-metal-core
  - Hardware state reporting
  - Command handling
  - Event propagation
- [ ] Define error handling conventions
  - Error categorization
  - Retry policies
  - Failure recovery
- [ ] Establish monitoring guidelines
  - Critical metrics
  - Performance indicators
  - Health checks
- [ ] Create integration test suite
  - End-to-end scenarios
  - Failure mode testing
  - Performance benchmarks

## Release Preparation
- [ ] Version tagging strategy
  - Semantic versioning
  - Hardware compatibility tags
  - Feature flags
- [ ] Release documentation
  - Installation guide
  - Configuration reference
  - Troubleshooting guide
- [ ] Migration guide from monolith
  - Step-by-step procedures
  - State preservation
  - Rollback procedures
- [ ] Changelog template
  - Feature additions
  - Bug fixes
  - Breaking changes

## Low Priority
- [ ] Optional features documentation
  - Advanced configurations
  - Custom sensor support
  - Alternative hardware
- [ ] Performance tuning guide
  - Optimization techniques
  - Resource management
  - Bottleneck analysis
- [ ] Hardware compatibility matrix
  - Tested configurations
  - Known issues
  - Workarounds
- [ ] Advanced simulation features
  - Complex fault injection
  - Environmental simulation
  - Performance modeling