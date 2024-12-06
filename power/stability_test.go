package power

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestStabilityMonitor(t *testing.T) {
	// Create mock power manager
	mockManager := &Manager{
		state: PowerState{
			Voltage:         5.0,
			PowerConsumption: 5.0,
			AvailablePower: map[PowerSource]bool{
				MainPower: true,
			},
			CurrentSource: MainPower,
		},
	}

	// Track stability events
	var events []StabilityEvent
	var eventsMux sync.Mutex

	// Create stability monitor
	monitor := newStabilityMonitor(StabilityConfig{
		SampleWindow:     100,
		SampleInterval:   10 * time.Millisecond,
		RippleThreshold:  0.5,
		CurrentThreshold: 2.0,
		OnStabilityEvent: func(event StabilityEvent) {
			eventsMux.Lock()
			events = append(events, event)
			eventsMux.Unlock()
		},
	})

	// Start monitoring
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go func() {
		if err := monitor.Monitor(ctx, mockManager); err != nil && err != context.DeadlineExceeded {
			t.Errorf("Monitor failed: %v", err)
		}
	}()

	// Test voltage stability detection
	t.Run("Voltage Stability", func(t *testing.T) {
		// Simulate voltage fluctuation
		mockManager.state.Voltage = 5.0
		time.Sleep(20 * time.Millisecond)
		mockManager.state.Voltage = 4.2
		time.Sleep(20 * time.Millisecond)
		mockManager.state.Voltage = 5.0
		time.Sleep(20 * time.Millisecond)

		metrics := monitor.GetMetrics()
		if metrics.VoltageRipple < 0.5 {
			t.Error("Expected to detect voltage ripple")
		}

		eventsMux.Lock()
		hasRippleEvent := false
		for _, event := range events {
			if event.Type == EventVoltageRipple {
				hasRippleEvent = true
				break
			}
		}
		eventsMux.Unlock()

		if !hasRippleEvent {
			t.Error("Expected voltage ripple event")
		}
	})

	// Test power cycling detection
	t.Run("Power Cycling", func(t *testing.T) {
		initialCycles := monitor.GetMetrics().PowerCycles

		// Simulate power loss and restore
		mockManager.state.AvailablePower[MainPower] = false
		time.Sleep(20 * time.Millisecond)
		mockManager.state.AvailablePower[MainPower] = true
		time.Sleep(20 * time.Millisecond)

		metrics := monitor.GetMetrics()
		if metrics.PowerCycles <= initialCycles {
			t.Error("Expected to detect power cycle")
		}
		if metrics.LastCycleDuration == 0 {
			t.Error("Expected non-zero cycle duration")
		}

		eventsMux.Lock()
		hasCycleEvent := false
		for _, event := range events {
			if event.Type == EventPowerCycle {
				hasCycleEvent = true
				break
			}
		}
		eventsMux.Unlock()

		if !hasCycleEvent {
			t.Error("Expected power cycle event")
		}
	})

	// Test current spike detection
	t.Run("Current Spikes", func(t *testing.T) {
		initialSpikes := monitor.GetMetrics().CurrentSpikes

		// Simulate current spike
		mockManager.state.PowerConsumption = 15.0 // 3A at 5V
		time.Sleep(20 * time.Millisecond)
		mockManager.state.PowerConsumption = 5.0  // Return to normal
		time.Sleep(20 * time.Millisecond)

		metrics := monitor.GetMetrics()
		if metrics.CurrentSpikes <= initialSpikes {
			t.Error("Expected to detect current spike")
		}
		if metrics.MaxCurrentSpike < 3.0 {
			t.Error("Expected to record max current spike")
		}

		eventsMux.Lock()
		hasSpikeEvent := false
		for _, event := range events {
			if event.Type == EventCurrentSpike {
				hasSpikeEvent = true
				break
			}
		}
		eventsMux.Unlock()

		if !hasSpikeEvent {
			t.Error("Expected current spike event")
		}
	})

	// Test source failover detection
	t.Run("Source Failover", func(t *testing.T) {
		// Simulate power source change
		mockManager.state.CurrentSource = BatteryPower
		time.Sleep(20 * time.Millisecond)

		eventsMux.Lock()
		hasFailoverEvent := false
		for _, event := range events {
			if event.Type == EventSourceFailover {
				hasFailoverEvent = true
				break
			}
		}
		eventsMux.Unlock()

		if !hasFailoverEvent {
			t.Error("Expected source failover event")
		}
	})
}
