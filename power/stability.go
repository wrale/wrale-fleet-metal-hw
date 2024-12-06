package power

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// stabilityMonitor handles power quality monitoring
type stabilityMonitor struct {
	mux  sync.RWMutex
	cfg  StabilityConfig
	
	// Circular buffer for voltage samples
	samples []float64
	head    int
	count   int
	
	// State tracking
	lastVoltage   float64
	lastSource    PowerSource
	cycleStart    time.Time
	cycleInProgress bool
	
	// Running statistics
	metrics StabilityMetrics
}

func newStabilityMonitor(cfg StabilityConfig) *stabilityMonitor {
	// Set reasonable defaults
	if cfg.SampleWindow == 0 {
		cfg.SampleWindow = 1000 // Keep last 1000 samples
	}
	if cfg.SampleInterval == 0 {
		cfg.SampleInterval = defaultStabilityInterval
	}
	if cfg.RippleThreshold == 0 {
		cfg.RippleThreshold = criticalRipple
	}
	if cfg.CurrentThreshold == 0 {
		cfg.CurrentThreshold = criticalCurrent
	}
	
	return &stabilityMonitor{
		cfg:     cfg,
		samples: make([]float64, cfg.SampleWindow),
	}
}

// Monitor starts stability monitoring in the background
func (s *stabilityMonitor) Monitor(ctx context.Context, m *Manager) error {
	ticker := time.NewTicker(s.cfg.SampleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.sample(m); err != nil {
				return fmt.Errorf("failed to sample power state: %w", err)
			}
		}
	}
}

// sample takes a power measurement and updates stability metrics
func (s *stabilityMonitor) sample(m *Manager) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state := m.GetState()
	voltage := state.Voltage
	current := state.PowerConsumption / voltage

	// Add voltage sample to circular buffer
	s.samples[s.head] = voltage
	s.head = (s.head + 1) % len(s.samples)
	if s.count < len(s.samples) {
		s.count++
	}

	// Update stability metrics
	metrics := &s.metrics
	metrics.UpdatedAt = time.Now()

	// Check for power source changes
	if state.CurrentSource != s.lastSource {
		if s.cfg.OnStabilityEvent != nil {
			s.cfg.OnStabilityEvent(StabilityEvent{
				Timestamp: time.Now(),
				Type:      EventSourceFailover,
				Source:    state.CurrentSource,
				Details:   fmt.Sprintf("Power source changed from %s to %s", s.lastSource, state.CurrentSource),
			})
		}
		s.lastSource = state.CurrentSource
	}

	// Detect power cycles
	powerAvailable := false
	for _, available := range state.AvailablePower {
		if available {
			powerAvailable = true
			break
		}
	}

	if !powerAvailable && !s.cycleInProgress {
		s.cycleInProgress = true
		s.cycleStart = time.Now()
	} else if powerAvailable && s.cycleInProgress {
		s.cycleInProgress = false
		duration := time.Since(s.cycleStart)
		metrics.PowerCycles++
		metrics.LastCycleDuration = duration

		if s.cfg.OnStabilityEvent != nil {
			s.cfg.OnStabilityEvent(StabilityEvent{
				Timestamp: time.Now(),
				Type:      EventPowerCycle,
				Reading:   float64(duration.Milliseconds()),
				Details:   fmt.Sprintf("Power cycle detected lasting %v", duration),
			})
		}
	}

	// Calculate voltage stability metrics
	if s.count > 0 {
		min, max := s.samples[0], s.samples[0]
		sum := 0.0
		for i := 0; i < s.count; i++ {
			v := s.samples[i]
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
			sum += v
		}

		metrics.MinVoltage = min
		metrics.MaxVoltage = max
		metrics.VoltageRipple = max - min
		metrics.AverageVoltage = sum / float64(s.count)

		// Check for excessive ripple
		if metrics.VoltageRipple > s.cfg.RippleThreshold {
			if s.cfg.OnStabilityEvent != nil {
				s.cfg.OnStabilityEvent(StabilityEvent{
					Timestamp: time.Now(),
					Type:      EventVoltageRipple,
					Reading:   metrics.VoltageRipple,
					Threshold: s.cfg.RippleThreshold,
					Details:   fmt.Sprintf("Voltage ripple %.2fV exceeds threshold %.2fV", metrics.VoltageRipple, s.cfg.RippleThreshold),
				})
			}
		}

		// Check for voltage sag
		if min < criticalVoltage {
			if s.cfg.OnStabilityEvent != nil {
				s.cfg.OnStabilityEvent(StabilityEvent{
					Timestamp: time.Now(),
					Type:      EventVoltageSag,
					Reading:   min,
					Threshold: criticalVoltage,
					Details:   fmt.Sprintf("Voltage sag to %.2fV below critical threshold %.2fV", min, criticalVoltage),
				})
			}
		}
	}

	// Check for current spikes
	if current > s.cfg.CurrentThreshold {
		metrics.CurrentSpikes++
		if current > metrics.MaxCurrentSpike {
			metrics.MaxCurrentSpike = current
		}

		if s.cfg.OnStabilityEvent != nil {
			s.cfg.OnStabilityEvent(StabilityEvent{
				Timestamp: time.Now(),
				Type:      EventCurrentSpike,
				Reading:   current,
				Threshold: s.cfg.CurrentThreshold,
				Details:   fmt.Sprintf("Current spike %.2fA exceeds threshold %.2fA", current, s.cfg.CurrentThreshold),
			})
		}
	}

	s.lastVoltage = voltage
	return nil
}

// GetMetrics returns a copy of current stability metrics
func (s *stabilityMonitor) GetMetrics() StabilityMetrics {
	s.mux.RLock()
	defer s.mux.RUnlock()
	
	// Return copy to maintain immutability
	metrics := s.metrics
	return metrics
}