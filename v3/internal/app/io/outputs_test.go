package io

import (
	"testing"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/stretchr/testify/assert"
)

func TestFormatFinalized(t *testing.T) {
	// Define a slice of structured test cases to ensure comprehensive coverage.
	tests := []struct {
		name                   string
		requestedEpoch         int
		latestFinalizedEpoch   int
		expectedFinalityParams model.FinalityParams
	}{
		{
			name:                   "Requested Epoch is Less Than Finalized (NOT_FINALIZED)",
			requestedEpoch:         100, // Requesting an older epoch
			latestFinalizedEpoch:   105,
			expectedFinalityParams: "not_finalized",
		},
		{
			name:                   "Requested Epoch Equals Finalized (FINALIZED)",
			requestedEpoch:         105,
			latestFinalizedEpoch:   105,
			expectedFinalityParams: "finalized",
		},
		{
			name:                   "Requested Epoch is Greater Than Finalized (FINALIZED)",
			requestedEpoch:         110, // Requesting an epoch newer than the latest finalized one
			latestFinalizedEpoch:   105,
			expectedFinalityParams: "finalized",
		},
		{
			name:                   "Zero Epochs Case (FINALIZED)",
			requestedEpoch:         0,
			latestFinalizedEpoch:   0,
			expectedFinalityParams: "finalized",
		},
		{
			name:                   "Edge Case: Requested is just 1 epoch less (NOT_FINALIZED)",
			requestedEpoch:         99,
			latestFinalizedEpoch:   100,
			expectedFinalityParams: "not_finalized",
		},
	}

	// Iterate over the test cases and run them as subtests.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function under test
			actual := FormatFinalized(tt.requestedEpoch, tt.latestFinalizedEpoch)

			// Assert that the actual result matches the expected result.
			// The message provides useful context on failure.
			assert.Equal(t, tt.expectedFinalityParams, actual, "FormatFinalized(%d, %d) failed", tt.requestedEpoch, tt.latestFinalizedEpoch)
		})
	}
}

// TestEncodeHexString tests the EncodeHexString function using a table-driven approach.
func TestEncodeHexString(t *testing.T) {
	// Define a slice of test cases.
	tests := []struct {
		name     string // Name of the test case
		input    []byte // Input byte slice
		expected string // Expected output string
	}{
		{
			name:     "Empty slice",
			input:    []byte{},
			expected: "0x", // Empty slice should result in "0x"
		},
		{
			name:     "Simple slice (1-byte)",
			input:    []byte{0x0A},
			expected: "0x0a", // '0A' -> '0a'
		},
		{
			name:     "Standard case (3 bytes)",
			input:    []byte{0x01, 0x02, 0x03},
			expected: "0x010203",
		},
		{
			name:     "Full range of byte values",
			input:    []byte{0xDE, 0xAD, 0xBE, 0xEF},
			expected: "0xdeadbeef",
		},
		{
			name:     "Bytes representing ASCII text",
			input:    []byte{'H', 'e', 'l', 'l', 'o'}, // ASCII: 48 65 6c 6c 6f
			expected: "0x48656c6c6f",
		},
	}

	// Iterate over the test cases and run them as subtests.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := EncodeHexString(tt.input)

			if actual != tt.expected {
				t.Errorf("EncodeHexString(%v) FAILED.\nExpected: %s\nActual:   %s",
					tt.input, tt.expected, actual)
			}
		})
	}
}
