# Hardware Testing Guide

## Physical Hardware Testing

### Required Equipment
- Raspberry Pi 3B+ or 5
- 5V/2A Power Supply
- Multimeter for power measurements
- Load banks for power testing
- Temperature monitoring equipment
- Security sensors (case, motion)
- PWM-compatible fan
- GPIO measurement tools

### Safety Precautions
- Never exceed 3.3V on GPIO pins
- Monitor temperature during stress tests
- Use ESD protection when handling hardware
- Keep proper ventilation during testing
- Monitor power consumption

### Test Sequence

1. Initial Setup
```bash
# Clone repository
git clone github.com/wrale/wrale-fleet-metal-hw
cd wrale-fleet-metal-hw

# Build test binary
go build -o hwtest cmd/hwtest/main.go

# Run hardware detection
./hwtest detect
```

2. GPIO Testing
```bash
# Test GPIO read/write
./hwtest gpio verify

# Test PWM functionality
./hwtest gpio pwm --pin 18
```

3. Power Testing
```bash
# Basic power validation
./hwtest power check

# Full load test (requires load banks)
./hwtest power load-test --duration 30m
```

4. Thermal Testing
```bash
# Temperature sensor validation
./hwtest thermal sensors

# Fan control test
./hwtest thermal fan --pin 18
```

5. Security Testing
```bash
# Test case sensors
./hwtest secure case

# Test motion detection
./hwtest secure motion
```

### Common Issues

1. GPIO Issues
- Check pin numbers against board diagram
- Verify 3.3V logic level
- Check pull-up/down configuration
- Monitor for floating inputs

2. Power Issues
- Verify power supply capacity
- Check voltage stability under load
- Monitor for voltage sags
- Validate power source switching

3. Thermal Issues
- Ensure proper thermal paste application
- Check fan connection and PWM signal
- Verify temperature sensor readings
- Monitor throttling behavior

4. Security Sensor Issues
- Check sensor orientation
- Verify trigger levels
- Test in various light conditions
- Validate debouncing settings

## Simulation Testing

### Setup Simulation Environment
```bash
# Start hardware simulation
make sim-start

# Run test suite with simulation
go test ./... -tags=simulation

# Stop simulation
make sim-stop
```

### Available Simulated Hardware
- GPIO pins with PWM support
- Power management simulation
- Temperature sensor simulation
- Security sensor simulation

### Simulation vs Real Hardware
1. Advantages
   - Reproducible tests
   - Fast execution
   - No hardware required
   - Easy fault injection

2. Limitations
   - No real timing
   - No physical interactions
   - Limited environmental factors
   - No true power behavior

### Adding Custom Tests
```go
// Example: Add custom GPIO test
func TestCustomGPIO(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping test in short mode")
    }
    
    ctrl := NewSimulatedGPIO()
    // Add test logic here
}
```

## Continuous Integration Testing

### GitHub Actions Setup
```yaml
name: Hardware Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
      - name: Run simulated tests
        run: |
          make sim-start
          go test ./... -tags=simulation
          make sim-stop
```

### Manual Testing Checklist
- [ ] All GPIO functions verified
- [ ] Power stability confirmed
- [ ] Temperature control tested
- [ ] Security features validated
- [ ] PWM operation checked
- [ ] Load tests completed
- [ ] Sensor calibration verified

## Test Coverage Requirements

### Core Requirements
1. GPIO Package: >90%
2. Power Package: >90%
3. Thermal Package: >85%
4. Security Package: >85%
5. Diagnostics: >80%

### Integration Requirements
1. Cross-package tests
2. Error handling paths
3. Resource cleanup
4. Concurrent operations

## Documentation Updates

After completing tests, update:
1. Update CHANGELOG.md
2. Document any hardware-specific issues
3. Update compatibility matrix
4. Note performance characteristics