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
