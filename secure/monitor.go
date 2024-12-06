package secure

import (
	"context"
	"fmt"
	"time"
)

// Monitor starts continuous security monitoring
func (m *Manager) Monitor(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.checkSecurity(ctx); err != nil {
				return fmt.Errorf("security check failed: %w", err)
			}
		}
	}
}

// checkSecurity performs a single security check
func (m *Manager) checkSecurity(ctx context.Context) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	// Check case sensor
	caseOpen, err := m.gpio.GetPinState(m.caseSensor)
	if err != nil {
		return fmt.Errorf("failed to check case sensor: %w", err)
	}

	// Check motion sensor
	motion, err := m.gpio.GetPinState(m.motionSensor)
	if err != nil {
		return fmt.Errorf("failed to check motion sensor: %w", err)
	}

	// Check voltage sensor
	voltageOK, err := m.gpio.GetPinState(m.voltSensor)
	if err != nil {
		return fmt.Errorf("failed to check voltage sensor: %w", err)
	}

	// Update state
	newState := TamperState{
		CaseOpen:       caseOpen,
		MotionDetected: motion,
		VoltageNormal:  voltageOK,
		LastCheck:      time.Now(),
	}

	// Check for tamper conditions
	if caseOpen || motion || !voltageOK {
		if m.onTamper != nil {
			m.onTamper(newState)
		}

		// Log tamper event if store is available
		if m.stateStore != nil {
			eventDetails := map[string]interface{}{
				"case_open":       caseOpen,
				"motion_detected": motion,
				"voltage_normal":  voltageOK,
			}
			if err := m.stateStore.LogEvent(ctx, m.deviceID, "tamper_detected", eventDetails); err != nil {
				// Log but don't fail on event logging error
				fmt.Printf("Failed to log tamper event: %v", err)
			}
		}
	}

	// Persist state if store is available
	if m.stateStore != nil {
		if err := m.stateStore.SaveState(ctx, m.deviceID, newState); err != nil {
			// Log but don't fail on state persistence error
			fmt.Printf("Failed to persist security state: %v", err)
		}
	}

	m.state = newState
	return nil
}
