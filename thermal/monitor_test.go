package thermal

import "testing"

func TestMonitor(t *testing.T) {
	// Skip hardware tests in CI
	t.Skip("Skipping thermal tests in CI - requires hardware")
}
