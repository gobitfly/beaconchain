package io

import (
	"testing"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestAsNullable_WithValue tests the conversion when the input pointer is not nil.
func TestAsNullable_WithValue(t *testing.T) {
	// Arrange
	originalValue := 42
	ptr := &originalValue

	// Act
	result := AsNullable(ptr)

	// Assert
	// 1. Assert it is NOT null using assert.False
	assert.False(t, result.IsNull(), "The result should not be null for a non-nil pointer")

	// 2. Assert the value retrieved matches the original
	retrievedValue, err := result.Get()
	require.NoError(t, err, "Get() should return no error when the Nullable is not null")
	assert.Equal(t, originalValue, retrievedValue, "The retrieved value should match the original pointer value")
}

// TestAsNullable_WithNil tests the conversion when the input pointer is nil.
func TestAsNullable_WithNil(t *testing.T) {
	// Arrange
	var ptr *string // Nil pointer of type *string

	// Act
	result := AsNullable(ptr)

	// Assert
	// 1. Assert it IS null using assert.True
	assert.True(t, result.IsNull(), "The result should be null for a nil pointer")

	// 2. Assert getting the value returns an expected error
	retrievedValue, err := result.Get()
	assert.Error(t, err, "Get() should return ErrNullValue when the Nullable is null")

	// 3. Assert the retrieved value is the zero value for the type (e.g., empty string)
	var zeroValue string
	assert.Equal(t, zeroValue, retrievedValue, "The retrieved value should be the zero value for the type")
}

// TestAsNullable_WithStruct tests the function with a custom struct type.
func TestAsNullable_WithStruct(t *testing.T) {
	// Arrange
	type User struct {
		Name string
		Age  int
	}

	// Case 1: Non-nil pointer
	user := User{Name: "Alice", Age: 30}
	ptr := &user
	result := AsNullable(ptr)

	assert.False(t, result.IsNull(), "The struct result should not be null")
	retrievedUser, err := result.Get()
	require.NoError(t, err)
	assert.Equal(t, user, retrievedUser, "Struct value should be correctly retrieved")

	// Case 2: Nil pointer
	var nilPtr *User
	nilResult := AsNullable(nilPtr)

	assert.True(t, nilResult.IsNull(), "The nil struct result should be null")
	_, err = nilResult.Get()
	assert.Error(t, err, "Get() should return an error for a null nullable struct")
}

// TestAsNullableValue tests the conversion of a direct value to a non-null Nullable.
func TestAsNullableValue(t *testing.T) {
	// Arrange
	originalInt := 101
	originalString := "Test Value"

	// Case 1: Integer
	t.Run("with integer value", func(t *testing.T) {
		// Act
		result := AsNullableValue(originalInt)

		// Assert
		// 1. Assert it is NOT null
		assert.False(t, result.IsNull(), "The result should not be null for a direct value")

		// 2. Assert the value retrieved matches the original
		retrievedValue, err := result.Get()
		require.NoError(t, err, "Get() should return no error when the Nullable is not null")
		assert.Equal(t, originalInt, retrievedValue, "The retrieved integer value should match the original value")
	})

	// Case 2: String
	t.Run("with string value", func(t *testing.T) {
		// Act
		result := AsNullableValue(originalString)

		// Assert
		// 1. Assert it is NOT null
		assert.False(t, result.IsNull(), "The result should not be null for a direct string value")

		// 2. Assert the value retrieved matches the original
		retrievedValue, err := result.Get()
		require.NoError(t, err, "Get() should return no error when the Nullable is not null")
		assert.Equal(t, originalString, retrievedValue, "The retrieved string value should match the original value")
	})
}

func TestAsWithdrawalCredential(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected model.WithdrawalCredential
	}{
		{
			name:  "Prefix 0x00 - BLS Credential",
			input: "0x00d8cfc9e2ad8166ce4b610373d3ccba57ceaddea3498eb8ec3ddb90202d87b5",
			expected: model.WithdrawalCredential{
				Address:    "", // Should be empty for 0x00
				Credential: "0x00d8cfc9e2ad8166ce4b610373d3ccba57ceaddea3498eb8ec3ddb90202d87b5",
				Prefix:     model.N0x00,
				Type:       model.Bls,
			},
		},
		{
			name:  "Prefix 0x01 - Execution Layer Address Credential",
			input: "0x0100000000000000000000005fdcb78ca9a1164c13428e5fc9582c8c48dab69f",
			expected: model.WithdrawalCredential{
				Address:    "0x5fdcb78ca9a1164c13428e5fc9582c8c48dab69f",
				Credential: "0x0100000000000000000000005fdcb78ca9a1164c13428e5fc9582c8c48dab69f",
				Prefix:     model.N0x01,
				Type:       model.ExecutionAddress,
			},
		},
		{
			name:  "Prefix 0x02 - Execution Layer Address Credential",
			input: "0x0200000000000000000000005fdcb78ca9a1164c13428e5fc9582c8c48dab69f",
			expected: model.WithdrawalCredential{
				Address:    "0x5fdcb78ca9a1164c13428e5fc9582c8c48dab69f",
				Credential: "0x0200000000000000000000005fdcb78ca9a1164c13428e5fc9582c8c48dab69f",
				Prefix:     model.N0x02,
				Type:       model.ExecutionAddress,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the input credential with the desired prefix byte
			inputCred, err := DecodeHexString(tt.input)
			require.NoError(t, err, "Failed to decode input hex string")
			result, err := AsWithdrawalCredential(inputCred)
			require.NoError(t, err, "AsWithdrawalCredential() returned an unexpected error")
			assert.Equal(t, tt.expected, result, "AsWithdrawalCredential() result mismatch")
		})
	}
}
