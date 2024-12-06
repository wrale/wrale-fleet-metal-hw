package gpio

import (
	"context"
	"sync"
	"testing"
	"time"

	"periph.io/x/conn/v3/gpio"
)

// mockInterruptPin mocks a pin with interrupt capabilities
type mockInterruptPin struct {
	mux   sync.RWMutex
	state bool
	edge  gpio.Edge
}

func (m *mockInterruptPin) In(pull gpio.Pull, edge gpio.Edge) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.edge = edge
	return nil
}

func (m *mockInterruptPin) Out(l gpio.Level) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.state = l == gpio.High
	return nil
}

func (m *mockInterruptPin) Read() gpio.Level {
	m.mux.RLock()
	defer m.mux.RUnlock()
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

func TestInterrupts(t *testing.T) {
	// Create controller
	ctrl, err := New()
	if err != nil {
		t.Fatalf("Failed to create GPIO controller: %v", err)
	}

	// Setup test pin
	pin := &mockInterruptPin{}
	pinName := "test_pin"
	if err := ctrl.ConfigurePin(pinName, pin, gpio.PullUp); err != nil {
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

	// Start monitoring
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		if err := ctrl.Monitor(ctx); err != nil && err != context.DeadlineExceeded {
			t.Errorf("Monitor failed: %v", err)
		}
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

	// Test debouncing
	t.Run("Debounce", func(t *testing.T) {
		// Reset count
		interruptMux.Lock()
		interruptCount = 0
		interruptMux.Unlock()

		// Trigger multiple interrupts rapidly
		for i := 0; i < 5; i++ {
			pin.Out(gpio.High)
			pin.Out(gpio.Low)
		}
		time.Sleep(20 * time.Millisecond)

		// Check if debouncing worked
		interruptMux.RLock()
		count := interruptCount
		interruptMux.RUnlock()

		if count > 2 { // Allow for some timing variation
			t.Errorf("Debouncing failed: got %d interrupts", count)
		}
	})

	// Test disable
	t.Run("Disable", func(t *testing.T) {
		// Disable interrupts
		if err := ctrl.DisableInterrupt(pinName); err != nil {
			t.Errorf("Failed to disable interrupt: %v", err)
		}

		// Reset count
		interruptMux.Lock()
		interruptCount = 0
		interruptMux.Unlock()

		// Try to trigger interrupt
		pin.Out(gpio.High)
		time.Sleep(20 * time.Millisecond)

		// Check that handler wasn't called
		interruptMux.RLock()
		count := interruptCount
		interruptMux.RUnlock()

		if count > 0 {
			t.Error("Interrupt handler called after disable")
		}
	})
}