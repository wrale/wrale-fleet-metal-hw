package gpio

import (
	"context"
	"sync"
	"testing"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
)

// mockInterruptPin mocks a pin with interrupt capabilities
type mockInterruptPin struct {
	sync.RWMutex  // embedded mutex
	state bool
	edge  gpio.Edge
	pullState gpio.Pull
}

func (m *mockInterruptPin) In(pull gpio.Pull, edge gpio.Edge) error {
	m.Lock()
	defer m.Unlock()
	m.edge = edge
	m.pullState = pull
	return nil
}

func (m *mockInterruptPin) Out(l gpio.Level) error {
	m.Lock()
	defer m.Unlock()
	m.state = l == gpio.High
	return nil
}

func (m *mockInterruptPin) Read() gpio.Level {
	m.RLock()
	defer m.RUnlock()
	if m.state {
		return gpio.High
	}
	return gpio.Low
}

// Implement required PinIO methods
func (m *mockInterruptPin) String() string { return "mock_pin" }
func (m *mockInterruptPin) Name() string   { return "MOCK_PIN" }
func (m *mockInterruptPin) Number() int    { return 0 }
func (m *mockInterruptPin) Function() string { return "In/Out" }
func (m *mockInterruptPin) Halt() error    { return nil }
func (m *mockInterruptPin) DefaultPull() gpio.Pull { return gpio.Float }
func (m *mockInterruptPin) PWM(duty gpio.Duty, f physic.Frequency) error { return nil }
func (m *mockInterruptPin) Pull() gpio.Pull { return m.pullState }
func (m *mockInterruptPin) WaitForEdge(timeout time.Duration) bool { return true }

func TestInterrupts(t *testing.T) {
	// Create controller in simulation mode
	ctrl, err := New(WithSimulation())
	if err != nil {
		t.Fatalf("Failed to create GPIO controller: %v", err)
	}

	// Setup test pin
	pin := &mockInterruptPin{}
	pinName := "test_pin"
	if err := ctrl.ConfigurePin(pinName, pin, gpio.Float); err != nil {
		t.Fatalf("Failed to configure pin: %v", err)
	}

	// Track interrupt calls
	var (
		interruptCount int
		interruptMux   sync.RWMutex
	)

	// Enable interrupt
	err = ctrl.EnableInterrupt(pinName, InterruptConfig{
		Edge:         Both,
		DebounceTime: 10 * time.Millisecond,
		Handler: func(p string, state bool) {
			interruptMux.Lock()
			interruptCount++
			interruptMux.Unlock()
		},
	})
	if err != nil {
		t.Fatalf("Failed to enable interrupt: %v", err)
	}

	// Use buffered channel to prevent potential deadlock
	monitorDone := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		monitorDone <- ctrl.Monitor(ctx)
	}()

	// Test interrupt triggering
	t.Run("Basic Interrupt", func(t *testing.T) {
		// Trigger interrupt
		if err := pin.Out(gpio.High); err != nil {
			t.Errorf("Failed to set pin high: %v", err)
		}
		time.Sleep(20 * time.Millisecond)

		// Check if handler was called
		interruptMux.RLock()
		count := interruptCount
		interruptMux.RUnlock()

		if count == 0 {
			t.Error("Interrupt handler not called")
		}
	})

	// Wait for monitor to complete
	select {
	case err := <-monitorDone:
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("Monitor failed: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Monitor did not complete in time")
	}
}