package secure

import (
	"context"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

func TestSecurityManager(t *testing.T) {
	// Skip hardware test in CI
	t.Skip("Skipping security test in CI - requires hardware")
}