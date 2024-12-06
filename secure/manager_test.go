package secure

import "testing"

func TestSecurityManager(t *testing.T) {
	// Skip hardware test in CI
	t.Skip("Skipping security test in CI - requires hardware")
}