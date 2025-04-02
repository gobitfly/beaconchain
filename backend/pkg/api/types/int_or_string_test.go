package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T {
	return &v
}
func TestIntOrStringUnmarshalJSON_ValidCases(t *testing.T) {
	validTests := []struct {
		name          string
		input         string
		expectedValue IntOrString
	}{
		{
			name:          "Valid Integer",
			input:         `123`,
			expectedValue: IntOrString{IntValue: ptr(uint64(123))},
		},
		{
			name:          "Valid String Number",
			input:         `"456"`,
			expectedValue: IntOrString{IntValue: ptr(uint64(456))},
		},
		{
			name:          "Valid String",
			input:         `"hello"`,
			expectedValue: IntOrString{StrValue: ptr("hello")},
		},
		{
			name:          "Valid String with Leading/Trailing Spaces",
			input:         `"  hello  "`,
			expectedValue: IntOrString{StrValue: ptr("hello")},
		},
		{
			name:          "Valid Number with Leading/Trailing Spaces",
			input:         `" 789 "`,
			expectedValue: IntOrString{IntValue: ptr(uint64(789))},
		},
		{
			name:          "String with Non-Numeric Content",
			input:         `"abc123"`,
			expectedValue: IntOrString{StrValue: ptr("abc123")},
		},
		{
			name:          "Empty String",
			input:         `""`,
			expectedValue: IntOrString{StrValue: ptr("")},
		},
	}

	for _, test := range validTests {
		t.Run(test.name, func(t *testing.T) {
			var value IntOrString
			err := json.Unmarshal([]byte(test.input), &value)

			assert.NoError(t, err)
			assert.True(t, value.IntValue != nil || value.StrValue != nil)
			assert.False(t, value.IntValue != nil && value.StrValue != nil)

			assert.Equal(t, test.expectedValue.IntValue, value.IntValue)
			assert.Equal(t, test.expectedValue.StrValue, value.StrValue)
		})
	}
}

func TestIntOrStringUnmarshalJSON_ErrorCases(t *testing.T) {
	errorTests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "Invalid JSON Format",
			input:         `{}`,
			expectedError: "failed to unmarshal IntOrString from json: {}",
		},
		{
			name:          "Boolean Value",
			input:         `true`,
			expectedError: "failed to unmarshal IntOrString from json: true",
		},
		{
			name:          "Random Value",
			input:         `a`,
			expectedError: "invalid character 'a' looking for beginning of value",
		},
		{
			name:          "Null Value",
			input:         `null`,
			expectedError: "null value not allowed",
		},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			var value IntOrString
			err := json.Unmarshal([]byte(test.input), &value)
			assert.True(t, value.IntValue == nil && value.StrValue == nil)
			assert.ErrorContains(t, err, test.expectedError)
		})
	}
}
