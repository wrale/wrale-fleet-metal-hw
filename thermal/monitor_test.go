package thermal

import (
	"context"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

func TestMonitor(t *testing.T) {
	// Skip hardware tests in CI
	t.Skip("Skipping thermal tests in CI - requires hardware")
}