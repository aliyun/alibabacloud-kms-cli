package kms

import (
	"testing"
)

func TestGetWaitTimeExponential(t *testing.T) {
	tests := []struct {
		retryCount int
		expected   int64
	}{
		{0, 200},
		{1, 400},
		{2, 800},
		{3, 1600},
		{4, 3200},
	}

	for _, tt := range tests {
		got := getWaitTimeExponential(tt.retryCount)
		if got != tt.expected {
			t.Errorf("getWaitTimeExponential(%d) = %d, want %d", tt.retryCount, got, tt.expected)
		}
	}
}

func TestGetWaitTimeExponential_NegativeRetryCount(t *testing.T) {
	// retryCount = -1 should result in a very small value (200 * 0.5)
	got := getWaitTimeExponential(-1)
	if got != 100 {
		t.Errorf("getWaitTimeExponential(-1) = %d, want 100", got)
	}
}
