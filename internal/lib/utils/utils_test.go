package utils

import (
	"testing"
	"time"
)

func TestTimeNow(t *testing.T) {
	timestamp := TimeNow()

	// Assertions
	if timestamp == "" {
		t.Error("Expected non-empty timestamp")
	}

	if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
		t.Errorf("Expected valid RFC3339 timestamp, got '%s': %v", timestamp, err)
	}
}
