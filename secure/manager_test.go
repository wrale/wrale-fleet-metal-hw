package secure

import (
	"context"
	"sync"
	"testing"
	"time"
)

// mockGPIO implements a mock GPIO controller for testing
type mockGPIO struct {
	mux       sync.RWMutex
	pinStates map[string]bool
}

func newMockGPIO() *mockGPIO {
	return &mockGPIO{
		pinStates: make(map[string]bool),
	}
}

func (m *mockGPIO) ConfigurePin(name string, _ interface{}) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.pinStates[name] = false
	return nil
}

func (m *mockGPIO) SetPinState(name string, high bool) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.pinStates[name] = high
	return nil
}

func (m *mockGPIO) GetPinState(name string) (bool, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.pinStates[name], nil
}

// mockStateStore implements a mock state persistence store for testing
type mockStateStore struct {
	mux    sync.RWMutex
	states map[string]TamperState
	events []Event
}

func newMockStateStore() *mockStateStore {
	return &mockStateStore{
		states: make(map[string]TamperState),
		events: make([]Event, 0),
	}
}

func (m *mockStateStore) SaveState(ctx context.Context, deviceID string, state TamperState) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.states[deviceID] = state
	return nil
}

func (m *mockStateStore) LoadState(ctx context.Context, deviceID string) (TamperState, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	if state, exists := m.states[deviceID]; exists {
		return state, nil
	}
	return TamperState{}, nil
}

func (m *mockStateStore) LogEvent(ctx context.Context, deviceID string, eventType string, details interface{}) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.events = append(m.events, Event{
		DeviceID:  deviceID,
		Type:      eventType,
		Timestamp: time.Now(),
		Details:   details,
	})
	return nil
}

func TestSecurityManager(t *testing.T) {
	// Setup mocks
	mockGPIO := newMockGPIO()
	mockStore := newMockStateStore()

	// Setup test pins
	caseSensor := "case_sensor"
	motionSensor := "motion_sensor"
	voltSensor := "volt_sensor"
	deviceID := "test_device_001"

	mockGPIO.ConfigurePin(caseSensor, nil)
	mockGPIO.ConfigurePin(motionSensor, nil)
	mockGPIO.ConfigurePin(voltSensor, nil)

	// Create security manager
	var tamperDetected bool
	manager, err := New(Config{
		GPIO:          mockGPIO,
		CaseSensor:    caseSensor,
		MotionSensor:  motionSensor,
		VoltageSensor: voltSensor,
		DeviceID:      deviceID,
		StateStore:    mockStore,
		OnTamper: func(state TamperState) {
			tamperDetected = true
		},
	})

	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Start monitoring in background
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go func() {
		if err := manager.Monitor(ctx); err != nil && err != context.DeadlineExceeded {
			t.Errorf("Monitor failed: %v", err)
		}
	}()

	// Test initial state (all normal)
	state := manager.GetState()
	if state.CaseOpen || state.MotionDetected || !state.VoltageNormal {
		t.Error("Unexpected initial state")
	}

	// Simulate case tamper
	mockGPIO.SetPinState(caseSensor, true)
	time.Sleep(200 * time.Millisecond)

	state = manager.GetState()
	if !state.CaseOpen {
		t.Error("Case tamper not detected")
	}
	if !tamperDetected {
		t.Error("Tamper callback not triggered")
	}

	// Verify state was persisted
	storedState, err := mockStore.LoadState(ctx, deviceID)
	if err != nil {
		t.Errorf("Failed to load persisted state: %v", err)
	}
	if !storedState.CaseOpen {
		t.Error("Persisted state does not reflect tamper")
	}

	// Verify event was logged
	if len(mockStore.events) == 0 {
		t.Error("No security events were logged")
	} else {
		event := mockStore.events[0]
		if event.Type != "tamper_detected" {
			t.Errorf("Unexpected event type: %s", event.Type)
		}
		if event.DeviceID != deviceID {
			t.Errorf("Unexpected device ID in event: %s", event.DeviceID)
		}
	}

	// Wait for monitoring to stop
	<-ctx.Done()
}