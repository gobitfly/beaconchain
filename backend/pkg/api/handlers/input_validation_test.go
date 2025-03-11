package handlers

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error(t *testing.T) {
	v := validationError{
		"field1": "must be a valid email",
		"field2": "cannot be empty",
	}

	err := v.AsError()

	assert.Error(t, err)
	assert.ErrorContains(t, err, "field1: must be a valid email")
	assert.ErrorContains(t, err, "field2: cannot be empty")
}

func TestValidationError_Add_NewError(t *testing.T) {
	var v validationError

	v.add("field1", "must be a valid email")

	assert.Len(t, v, 1)
	assert.Equal(t, "must be a valid email", v["field1"])
}

func TestValidationError_Add_MultipleErrorsForSameField(t *testing.T) {
	var v validationError

	v.add("field1", "must be a valid email")
	v.add("field1", "cannot be empty")

	assert.Len(t, v, 1)
	assert.Equal(t, "must be a valid email; cannot be empty", v["field1"])
}

func TestValidationError_Add_MultipleFields(t *testing.T) {
	var v validationError

	v.add("field1", "must be a valid email")
	v.add("field2", "cannot be empty")

	assert.Len(t, v, 2)
	assert.Equal(t, "must be a valid email", v["field1"])
	assert.Equal(t, "cannot be empty", v["field2"])
}

func TestValidationError_HasErrors_Empty(t *testing.T) {
	var v validationError
	err := v.AsError()
	assert.Nil(t, v)
	assert.NoError(t, err)
	assert.False(t, v.hasErrors())
	v = validationError{}
	err = v.AsError()
	assert.NotNil(t, v)
	assert.NoError(t, err)
	assert.False(t, v.hasErrors())
}

func TestValidationError_HasErrors_NonEmpty(t *testing.T) {
	var v validationError

	v.add("field1", "must be a valid email")

	assert.True(t, v.hasErrors())
}

type validationTestCase[T any] struct {
	name     string
	param    string
	expected T
	errMsg   string // if not empty, expect an error with this message
}

type validationFunc[T any] func(v *validationError, testCase validationTestCase[T]) T

func runValidationTests[T any](t *testing.T, testCases []validationTestCase[T], validate validationFunc[T]) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := validate(&v, tt)
			err := v.AsError()

			if tt.errMsg != "" {
				assert.Error(t, err, "Expected an error but got none")
				assert.ErrorContains(t, err, tt.errMsg)
			} else {
				assert.Nil(t, err, "Expected no errors but found some")
				assert.Equal(t, tt.expected, result, "Expected correct value")
			}
		})
	}
}

func TestCheckRegex(t *testing.T) {
	tests := []struct {
		name      string
		regex     *regexp.Regexp
		param     string
		expectErr bool
	}{
		{
			name:      "Valid input (letters only)",
			regex:     regexp.MustCompile(`^[a-zA-Z]+$`),
			param:     "ValidInput",
			expectErr: false,
		},
		{
			name:      "Invalid input (numbers included)",
			regex:     regexp.MustCompile(`^[a-zA-Z]+$`),
			param:     "123Invalid",
			expectErr: true,
		},
		{
			name:      "Invalid input (special characters)",
			regex:     regexp.MustCompile(`^[a-zA-Z]+$`),
			param:     "Hello@123",
			expectErr: true,
		},
		{
			name:      "Empty input",
			regex:     regexp.MustCompile(`^[a-zA-Z]+$`),
			param:     "",
			expectErr: true,
		},
		{
			name:      "Valid email format",
			regex:     regexp.MustCompile(`^[\w\.-]+@[\w\.-]+\.\w+$`),
			param:     "test@example.com",
			expectErr: false,
		},
		{
			name:      "Invalid email format",
			regex:     regexp.MustCompile(`^[\w\.-]+@[\w\.-]+\.\w+$`),
			param:     "invalid-email",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkRegex(tt.regex, tt.param, "param")
			err := v.AsError()

			assert.Equal(t, tt.param, result, "Expected input to be returned unchanged")

			if tt.expectErr {
				assert.Error(t, err, "Expected an error but got none")
				assert.ErrorContains(t, err, fmt.Sprintf(`given value '%s' has incorrect format`, tt.param))
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
			}
		})
	}
}

func TestCheckLength(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		minLength int
		errMsg    string
	}{
		{
			name:      "Valid length",
			param:     "ValidName",
			minLength: 3,
		},
		{
			name:      "Too short",
			param:     "Yo",
			minLength: 3,
			errMsg:    "given value 'Yo' is too short, minimum length is 3",
		},
		{
			name:      "Too long",
			param:     strings.Repeat("a", maxNameLength*2),
			minLength: 3,
			errMsg:    fmt.Sprintf("given value '%s' is too long, maximum length is %d", strings.Repeat("a", maxNameLength*2), maxNameLength),
		},
		{
			name:      "Exactly min length",
			param:     "Min",
			minLength: 3,
		},
		{
			name:      "Exactly max length",
			param:     strings.Repeat("a", maxNameLength),
			minLength: 3,
		},
		{
			name:      "Empty string",
			param:     "",
			minLength: 1,
			errMsg:    "given value '' is too short, minimum length is 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkLength(tt.param, "param", tt.minLength)
			err := v.AsError()

			assert.Equal(t, tt.param, result, "Expected input to be returned unchanged")

			if tt.errMsg != "" {
				assert.Error(t, err, "Expected an error but got none")
				assert.ErrorContains(t, err, tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
			}
		})
	}
}

func TestCheckName(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		minLength int
		expected  string
		errMsg    string
	}{
		{
			name:      "Valid name",
			param:     "Valid_Name-123",
			expected:  "Valid_Name-123",
			minLength: 3,
		},
		{
			name:      "Too short",
			param:     "Yo",
			minLength: 3,
			errMsg:    "given value 'Yo' is too short, minimum length is 3",
		},
		{
			name:      "Invalid characters",
			param:     "Invalid@Name",
			minLength: 3,
			errMsg:    "given value 'Invalid@Name' has incorrect format",
		},
		{
			name:      "Only spaces",
			param:     "   ",
			expected:  "   ",
			minLength: 3,
		},
		{
			name:      "Valid with space",
			param:     "John Doe",
			expected:  "John Doe",
			minLength: 3,
		},
		{
			name:      "Empty string",
			param:     "",
			minLength: 1,
			errMsg:    "given value '' is too short, minimum length is 1",
		},
		{
			name:      "Valid with dot",
			param:     "Dr. Smith",
			expected:  "Dr. Smith",
			minLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkName(tt.param, tt.minLength)
			err := v.AsError()

			if tt.errMsg != "" {
				assert.Error(t, err, "Expected an error but got none")
				assert.ErrorContains(t, err, tt.errMsg)
			} else {
				assert.Nil(t, err, "Expected no errors but found some")
				assert.Equal(t, tt.expected, result, "Expected correct value")
			}
		})
	}
}

func TestCheckNameNotEmpty(t *testing.T) {
	tests := []validationTestCase[string]{
		{
			name:     "Valid name",
			param:    "John",
			expected: "John",
		},
		{
			name:   "Empty string",
			param:  "",
			errMsg: "given value '' is too short, minimum length is 1",
		},
		{
			name:     "Only spaces",
			param:    "   ",
			expected: "   ",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[string]) string {
		return v.checkNameNotEmpty(tt.param)
	})
}

func TestCheckKeyNotEmpty(t *testing.T) {
	tests := []validationTestCase[string]{
		{
			name:     "Valid key",
			param:    "ValidKey123",
			expected: "ValidKey123",
		},
		{
			name:   "Empty string",
			param:  "",
			errMsg: "given value '' is too short, minimum length is 1",
		},
		{
			name:   "Invalid characters",
			param:  "Invalid@Key",
			errMsg: "given value 'Invalid@Key' has incorrect format",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[string]) string {
		return v.checkKeyNotEmpty(tt.param)
	})
}

func TestCheckEmail(t *testing.T) {
	tests := []validationTestCase[string]{
		{
			name:     "Valid email",
			param:    "test@example.com",
			expected: "test@example.com",
		},
		{
			name:     "Valid email with +",
			param:    "test+test@example.com",
			expected: "test+test@example.com",
		},
		{
			name:     "Valid email with uppercase gets converted to lowercase",
			param:    "Test@Example.COM",
			expected: "test@example.com",
		},
		{
			name:   "Missing @ symbol",
			param:  "invalid-email.com",
			errMsg: "given value 'invalid-email.com' has incorrect format",
		},
		{
			name:   "Missing domain",
			param:  "user@",
			errMsg: "given value 'user@' has incorrect format",
		},
		{
			name:   "Empty string",
			param:  "",
			errMsg: "given value '' has incorrect format",
		},
		{
			name:   "Invalid special characters",
			param:  "user@exa mple.com",
			errMsg: "given value 'user@exa mple.com' has incorrect format",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[string]) string {
		return v.checkEmail(tt.param)
	})
}

func TestCheckPassword(t *testing.T) {
	tests := []validationTestCase[string]{
		{
			name:     "Valid password (minimum length)",
			param:    "12345",
			expected: "12345",
		},
		{
			name:     "Valid password (longer)",
			param:    "securePassword123!",
			expected: "securePassword123!",
		},
		{
			name:   "Too short",
			param:  "1234",
			errMsg: "given value '1234' has incorrect format",
		},
		{
			name:   "Empty string",
			param:  "",
			errMsg: "given value '' has incorrect format",
		},
		{
			name:     "Whitespace only (should still pass since regex only enforces length)",
			param:    "     ",
			expected: "     ",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[string]) string {
		return v.checkPassword(tt.param)
	})
}

func TestCheckUserEmailToken(t *testing.T) {
	tests := []validationTestCase[string]{
		{
			name:     "Valid token (40 lowercase alphanumeric characters)",
			param:    "abcd1234abcd1234abcd1234abcd1234abcd1234",
			expected: "abcd1234abcd1234abcd1234abcd1234abcd1234",
		},
		{
			name:   "Too short",
			param:  "abcd1234abcd1234",
			errMsg: "given value 'abcd1234abcd1234' has incorrect format",
		},
		{
			name:   "Too long",
			param:  "abcd1234abcd1234abcd1234abcd1234abcd1234abcd",
			errMsg: "given value 'abcd1234abcd1234abcd1234abcd1234abcd1234abcd' has incorrect format",
		},
		{
			name:   "Contains uppercase letters",
			param:  "ABCD1234abcd1234abcd1234abcd1234abcd1234",
			errMsg: "given value 'ABCD1234abcd1234abcd1234abcd1234abcd1234' has incorrect format",
		},
		{
			name:   "Contains special characters",
			param:  "abcd1234abcd1234abcd1234abcd1234abcd12$%",
			errMsg: "given value 'abcd1234abcd1234abcd1234abcd1234abcd12$%' has incorrect format",
		},
		{
			name:   "Empty string",
			param:  "",
			errMsg: "given value '' has incorrect format",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[string]) string {
		return v.checkUserEmailToken(tt.param)
	})
}
func TestCheckValidatorDashboardPublicId(t *testing.T) {
	tests := []validationTestCase[types.VDBIdPublic]{
		{
			name:     "Valid public dashboard ID",
			param:    "v-12345678-1234-1234-1234-123456789abc",
			expected: types.VDBIdPublic("v-12345678-1234-1234-1234-123456789abc"),
		},
		{
			name:   "Invalid format (missing 'v-')",
			param:  "12345678-1234-1234-1234-123456789abc",
			errMsg: "given value '12345678-1234-1234-1234-123456789abc' has incorrect format",
		},
		{
			name:   "Invalid format (wrong character in UUID)",
			param:  "v-12345678-1234-1234-1234-123456789xyz",
			errMsg: "given value 'v-12345678-1234-1234-1234-123456789xyz' has incorrect format",
		},
		{
			name:   "Invalid format (too short)",
			param:  "v-12345678-1234-1234-1234-12345678",
			errMsg: "given value 'v-12345678-1234-1234-1234-12345678' has incorrect format",
		},
		{
			name:   "Empty string",
			param:  "",
			errMsg: "given value '' has incorrect format",
		},
	}
	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[types.VDBIdPublic]) types.VDBIdPublic {
		return v.checkValidatorDashboardPublicId(tt.param)
	})
}
func TestCheckAddress(t *testing.T) {
	tests := []validationTestCase[string]{
		{
			name:     "Valid address with 0x prefix",
			param:    "0x1234567890abcdef1234567890abcdef12345678",
			expected: "0x1234567890abcdef1234567890abcdef12345678",
		},
		{
			name:     "Valid address without 0x prefix",
			param:    "1234567890abcdef1234567890abcdef12345678",
			expected: "1234567890abcdef1234567890abcdef12345678",
		},
		{
			name:   "Invalid length (too short)",
			param:  "0x123456",
			errMsg: "given value '0x123456' has incorrect format",
		},
		{
			name:   "Invalid length (too long)",
			param:  "0x1234567890abcdef1234567890abcdef1234567890",
			errMsg: "given value '0x1234567890abcdef1234567890abcdef1234567890' has incorrect format",
		},
		{
			name:   "Invalid characters",
			param:  "0x1234567890abcdef1234567890abcdef1234567G",
			errMsg: "given value '0x1234567890abcdef1234567890abcdef1234567G' has incorrect format",
		},
		{
			name:   "Empty string",
			param:  "",
			errMsg: "given value '' has incorrect format",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[string]) string {
		return v.checkAddress(tt.param)
	})
}

func TestCheckInt(t *testing.T) {
	tests := []validationTestCase[int64]{
		{
			name:     "Valid positive integer",
			param:    "123",
			expected: 123,
		},
		{
			name:     "Valid negative integer",
			param:    "-456",
			expected: -456,
		},
		{
			name:     "Zero",
			param:    "0",
			expected: 0,
		},
		{
			name:   "Non-numeric string",
			param:  "abc",
			errMsg: "given value 'abc' is not an integer",
		},
		{
			name:   "Floating point number",
			param:  "3.14",
			errMsg: "given value '3.14' is not an integer",
		},
		{
			name:   "Empty string",
			param:  "",
			errMsg: "given value '' is not an integer",
		},
		{
			name:     "Large integer",
			param:    "9223372036854775807", // Max int64
			expected: math.MaxInt64,
		},
		{
			name:   "Too large for int64",
			param:  "9223372036854775808",
			errMsg: "given value '9223372036854775808' is not an integer",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[int64]) int64 {
		return v.checkInt(tt.param, "param")
	})
}

func TestCheckUint(t *testing.T) {
	tests := []validationTestCase[uint64]{
		{
			name:     "Valid positive integer",
			param:    "123",
			expected: 123,
		},
		{
			name:     "Zero",
			param:    "0",
			expected: 0,
		},
		{
			name:   "Negative number",
			param:  "-5",
			errMsg: "given value -5 is not a positive integer",
		},
		{
			name:   "Non-numeric string",
			param:  "abc",
			errMsg: "given value abc is not a positive integer",
		},
		{
			name:   "Floating point number",
			param:  "3.14",
			errMsg: "given value 3.14 is not a positive integer",
		},
		{
			name:   "Empty string",
			param:  "",
			errMsg: "given value  is not a positive integer",
		},
		{
			name:     "Max uint64",
			param:    "18446744073709551615", // Max uint64
			expected: math.MaxUint64,
		},
		{
			name:   "Too large for uint64",
			param:  "18446744073709551616",
			errMsg: "given value 18446744073709551616 is not a positive integer",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[uint64]) uint64 {
		return v.checkUint(tt.param, "param")
	})
}
func TestCheckWeiDecimal(t *testing.T) {
	tests := []validationTestCase[decimal.Decimal]{
		{
			name:     "Valid wei value",
			param:    "1000000000000000000",
			expected: decimal.RequireFromString("1000000000000000000"),
		},
		{
			name:     "Zero",
			param:    "0",
			expected: decimal.RequireFromString("0"),
		},
		{
			name:   "Negative number",
			param:  "-1",
			errMsg: "given value '-1' is not a wei string (must be positive integer)",
		},
		{
			name:   "Non-numeric string",
			param:  "abc",
			errMsg: "given value 'abc' is not a wei string (must be positive integer)",
		},
		{
			name:   "Floating point number",
			param:  "3.14",
			errMsg: "given value '3.14' is not a wei string (must be positive integer)",
		},
		{
			name:   "Empty string",
			param:  "",
			errMsg: "given value '' is not a wei string (must be positive integer)",
		},
		{
			name:     "Leading zeros (valid)",
			param:    "000123456789",
			expected: decimal.RequireFromString("123456789"),
		},
		{
			name:     "Very large wei value",
			param:    "115792089237316195423570985008687907853269984665640564039457584007913129639935", // max uint256
			expected: decimal.RequireFromString("115792089237316195423570985008687907853269984665640564039457584007913129639935"),
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[decimal.Decimal]) decimal.Decimal {
		return v.checkWeiDecimal(tt.param, "param")
	})
}

func TestCheckWeiMinMax(t *testing.T) {
	tests := []validationTestCase[decimal.Decimal]{
		{
			name:     "Valid value within range",
			param:    "50000",
			expected: decimal.RequireFromString("50000"),
		},
		{
			name:     "Exactly at minimum",
			param:    "1000",
			expected: decimal.RequireFromString("1000"),
		},
		{
			name:     "Exactly at maximum",
			param:    "100000",
			expected: decimal.RequireFromString("100000"),
		},
		{
			name:   "Below minimum",
			param:  "999",
			errMsg: "given value '999' is too small, minimum value is 1000",
		},
		{
			name:   "Above maximum",
			param:  "100001",
			errMsg: "given value '100001' is too large, maximum value is 100000",
		},
		{
			name:   "Invalid non-numeric input",
			param:  "abc",
			errMsg: "given value 'abc' is not a wei string (must be positive integer)",
		},
		{
			name:   "Empty string",
			param:  "",
			errMsg: "given value '' is not a wei string (must be positive integer)",
		},
		{
			name:   "Negative number",
			param:  "-5000",
			errMsg: "given value '-5000' is not a wei string (must be positive integer)",
		},
	}

	min := decimal.RequireFromString("1000")
	max := decimal.RequireFromString("100000")
	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[decimal.Decimal]) decimal.Decimal {
		return v.checkWeiMinMax(tt.param, "param", min, max)
	})
}

func TestCheckMinMax(t *testing.T) {
	tests := []struct {
		name      string
		param     int
		min       int
		max       int
		expectErr bool
		expected  int
		errMsg    string
	}{
		{
			name:      "Within range",
			param:     50,
			min:       10,
			max:       100,
			expectErr: false,
			expected:  50,
		},
		{
			name:      "Exactly at minimum",
			param:     10,
			min:       10,
			max:       100,
			expectErr: false,
			expected:  10,
		},
		{
			name:      "Exactly at maximum",
			param:     100,
			min:       10,
			max:       100,
			expectErr: false,
			expected:  100,
		},
		{
			name:      "Below minimum",
			param:     5,
			min:       10,
			max:       100,
			expectErr: true,
			expected:  5,
			errMsg:    "given value '5' is too small, minimum value is 10",
		},
		{
			name:      "Above maximum",
			param:     150,
			min:       10,
			max:       100,
			expectErr: true,
			expected:  150,
			errMsg:    "given value '150' is too large, maximum value is 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := checkMinMax(&v, tt.param, tt.min, tt.max, "param")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["param"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
				assert.Equal(t, tt.expected, result, "Expected correct value within range")
			}
		})
	}
}

func TestCheckUintMinMax(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		min       uint64
		max       uint64
		expectErr bool
		expected  uint64
		errMsg    string
	}{
		{
			name:      "Valid within range",
			param:     "50",
			min:       10,
			max:       100,
			expectErr: false,
			expected:  50,
		},
		{
			name:      "Exactly at minimum",
			param:     "10",
			min:       10,
			max:       100,
			expectErr: false,
			expected:  10,
		},
		{
			name:      "Exactly at maximum",
			param:     "100",
			min:       10,
			max:       100,
			expectErr: false,
			expected:  100,
		},
		{
			name:      "Below minimum",
			param:     "5",
			min:       10,
			max:       100,
			expectErr: true,
			expected:  5,
			errMsg:    "given value '5' is too small, minimum value is 10",
		},
		{
			name:      "Above maximum",
			param:     "150",
			min:       10,
			max:       100,
			expectErr: true,
			expected:  150,
			errMsg:    "given value '150' is too large, maximum value is 100",
		},
		{
			name:      "Empty string",
			param:     "",
			min:       10,
			max:       100,
			expectErr: true,
			errMsg:    "given value  is not a positive integer",
		},
		{
			name:      "Negative number (invalid for uint64)",
			param:     "-5",
			min:       10,
			max:       100,
			expectErr: true,
			errMsg:    "given value -5 is not a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkUintMinMax(tt.param, tt.min, tt.max, "param")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["param"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
				assert.Equal(t, tt.expected, result, "Expected correct uint64 within range")
			}
		})
	}
}

func TestSplitParameters(t *testing.T) {
	tests := []struct {
		name     string
		param    string
		delim    rune
		expected []string
	}{
		{
			name:     "Single value",
			param:    "value",
			delim:    ',',
			expected: []string{"value"},
		},
		{
			name:     "Multiple values with comma",
			param:    "one,two,three",
			delim:    ',',
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "Multiple values with spaces and comma",
			param:    " one , two , three ",
			delim:    ',',
			expected: []string{" one ", " two ", " three "},
		},
		{
			name:     "Empty string",
			param:    "",
			delim:    ',',
			expected: []string{},
		},
		{
			name:     "Only delimiters",
			param:    ",,,,",
			delim:    ',',
			expected: []string{},
		},
		{
			name:     "Mixed empty and non-empty values",
			param:    "one,,two,,three",
			delim:    ',',
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "Newline delimiter",
			param:    "line1\nline2\nline3",
			delim:    '\n',
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "Whitespace as delimiter",
			param:    "one two  three",
			delim:    ' ',
			expected: []string{"one", "two", "three"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitParameters(tt.param, tt.delim)

			assert.Equal(t, tt.expected, result, "Expected correct splitting of parameters")
		})
	}
}

func TestCheckAdConfigurationKeys(t *testing.T) {
	tests := []validationTestCase[[]string]{
		{
			name:     "Valid single key",
			param:    "validKey",
			expected: []string{"validKey"},
		},
		{
			name:     "Valid multiple keys",
			param:    "key1,key2,key3",
			expected: []string{"key1", "key2", "key3"},
		},
		{
			name:     "Valid keys with spaces",
			param:    " key1 , key2 , key3 ",
			expected: []string{"key1", "key2", "key3"},
		},
		{
			name:     "Empty string (should return empty slice)",
			param:    "",
			expected: []string{},
		},
		{
			name:   "Single invalid key",
			param:  "invalid@key",
			errMsg: "given value 'invalid@key' has incorrect format",
		},
		{
			name:   "One invalid key among valid ones",
			param:  "validKey,invalid@key,anotherValid",
			errMsg: "given value 'invalid@key' has incorrect format",
		},
		{
			name:   "Only invalid keys",
			param:  "invalid@key,$wrongKey,another#bad",
			errMsg: "given value 'invalid@key' has incorrect format",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[[]string]) []string {
		return v.checkAdConfigurationKeys(tt.param)
	})
}

func TestCheckBool(t *testing.T) {
	tests := []validationTestCase[bool]{
		{
			name:     "Valid true (lowercase)",
			param:    "true",
			expected: true,
		},
		{
			name:     "Valid false (lowercase)",
			param:    "false",
			expected: false,
		},
		{
			name:     "Valid true (uppercase T)",
			param:    "True",
			expected: true,
		},
		{
			name:     "Valid false (uppercase F)",
			param:    "False",
			expected: false,
		},
		{
			name:     "Valid true (1)",
			param:    "1",
			expected: true,
		},
		{
			name:     "Valid false (0)",
			param:    "0",
			expected: false,
		},
		{
			name:   "Invalid boolean string",
			param:  "yes",
			errMsg: "given value 'yes' is not a boolean",
		},
		{
			name:     "Empty string (should return false without error)",
			param:    "",
			expected: false,
		},
		{
			name:   "Random string",
			param:  "randomText",
			errMsg: "given value 'randomText' is not a boolean",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[bool]) bool {
		return v.checkBool(tt.param, "param")
	})
}

func TestCheckPrimaryDashboardId(t *testing.T) {
	tests := []validationTestCase[types.VDBIdPrimary]{
		{
			name:     "Valid dashboard ID",
			param:    "123",
			expected: types.VDBIdPrimary(123),
		},
		{
			name:     "Zero (valid ID)",
			param:    "0",
			expected: types.VDBIdPrimary(0),
		},
		{
			name:   "Negative number",
			param:  "-5",
			errMsg: "given value -5 is not a positive integer",
		},
		{
			name:   "Non-numeric string",
			param:  "abc",
			errMsg: "given value abc is not a positive integer",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[types.VDBIdPrimary]) types.VDBIdPrimary {
		return v.checkPrimaryDashboardId(tt.param)
	})
}

func TestCheckGroupId(t *testing.T) {
	tests := []struct {
		name       string
		param      string
		allowEmpty bool
		expected   int64
		errMsg     string
	}{
		{
			name:       "Valid group ID",
			param:      "123",
			allowEmpty: false,
			expected:   123,
		},
		{
			name:       "Zero (valid ID)",
			param:      "0",
			allowEmpty: false,
			expected:   0,
		},
		{
			name:       "Negative number",
			param:      "-1",
			allowEmpty: false,
			expected:   -1,
		},
		{
			name:       "Empty string with allowEmpty=true (should return AllGroups)",
			param:      "",
			allowEmpty: true,
			expected:   types.AllGroups,
		},
		{
			name:       "Empty string with allowEmpty=false (should fail)",
			param:      "",
			allowEmpty: false,
			errMsg:     "given value '' is not an integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkGroupId(tt.param, tt.allowEmpty)
			err := v.AsError()

			if tt.errMsg != "" {
				assert.Error(t, err, "Expected an error but got none")
				assert.ErrorContains(t, err, tt.errMsg)
			} else {
				assert.Nil(t, err, "Expected no errors but found some")
				assert.Equal(t, tt.expected, result, "Expected correct value")
			}
		})
	}
}

func TestCheckExistingGroupId(t *testing.T) {
	tests := []validationTestCase[uint64]{
		{
			name:     "Valid group ID",
			param:    "123",
			expected: 123,
		},
		{
			name:     "Zero (valid ID)",
			param:    "0",
			expected: 0,
		},
		{
			name:   "Negative number (invalid)",
			param:  "-5",
			errMsg: "given value -5 is not a positive integer",
		},
		{
			name:   "Empty string (invalid)",
			param:  "",
			errMsg: "given value  is not a positive integer",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[uint64]) uint64 {
		return v.checkExistingGroupId(tt.param)
	})
}

func TestParseGroupIdList(t *testing.T) {
	convertMock := func(id string, paramName string) int {
		num, _ := strconv.Atoi(id) // Simulating conversion function
		return num
	}

	tests := []struct {
		name     string
		param    string
		expected []int
	}{
		{
			name:     "Single valid ID",
			param:    "123",
			expected: []int{123},
		},
		{
			name:     "Multiple valid IDs",
			param:    "1,2,3",
			expected: []int{1, 2, 3},
		},
		{
			name:     "IDs with spaces",
			param:    " 1 , 2 , 3 ",
			expected: []int{1, 2, 3},
		},
		{
			name:     "Empty string (should return nil slice)",
			param:    "",
			expected: nil,
		},
		{
			name:     "delimiter only string (should return nil slice)",
			param:    ",",
			expected: nil,
		},
		{
			name:     "Extra commas",
			param:    ",1,,2,3,",
			expected: []int{1, 2, 3},
		},
		{
			name:     "Invalid numbers (should ignore but normally would return errors)",
			param:    "1,a,3",
			expected: []int{1, 0, 3}, // "a" converts to 0 due to strconv.Atoi ignoring errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGroupIdList(tt.param, convertMock)

			assert.Equal(t, tt.expected, result, "Expected correct parsing and conversion")
		})
	}
}

func TestCheckExistingGroupIdList(t *testing.T) {
	tests := []validationTestCase[[]uint64]{
		{
			name:     "Valid single ID",
			param:    "123",
			expected: []uint64{123},
		},
		{
			name:     "Multiple valid IDs",
			param:    "1,2,3",
			expected: []uint64{1, 2, 3},
		},
		{
			name:     "IDs with spaces",
			param:    " 1 , 2 , 3 ",
			expected: []uint64{1, 2, 3},
		},
		{
			name:     "Empty string (should return nil)",
			param:    "",
			expected: nil,
		},
		{
			name:   "Contains invalid ID (negative)",
			param:  "1,-2,3",
			errMsg: "given value -2 is not a positive integer",
		},
		{
			name:   "Contains non-numeric ID",
			param:  "1,abc,3",
			errMsg: "given value abc is not a positive integer",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[[]uint64]) []uint64 {
		return v.checkExistingGroupIdList(tt.param)
	})
}

func TestCheckGroupIdList(t *testing.T) {
	tests := []validationTestCase[[]int64]{
		{
			name:     "Valid single ID",
			param:    "123",
			expected: []int64{123},
		},
		{
			name:     "Multiple valid IDs",
			param:    "1,2,3",
			expected: []int64{1, 2, 3},
		},
		{
			name:     "IDs with spaces",
			param:    " 1 , 2 , 3 ",
			expected: []int64{1, 2, 3},
		},
		{
			name:     "Empty string (should return nil)",
			param:    "",
			expected: nil,
		},
		{
			name:   "Contains non-numeric ID",
			param:  "1,abc,3",
			errMsg: "given value 'abc' is not an integer",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[[]int64]) []int64 {
		return v.checkGroupIdList(tt.param)
	})
}

func TestCheckPagingParams(t *testing.T) {
	tests := []struct {
		name     string
		query    url.Values
		expected Paging
		errMsg   string
	}{
		{
			name: "Valid paging parameters",
			query: url.Values{
				"cursor": []string{"validCursor123"},
				"limit":  []string{"10"},
				"search": []string{"test"},
			},
			expected: Paging{
				cursor: "validCursor123",
				limit:  10,
				search: "test",
			},
		},
		{
			name: "Valid cursor, missing limit (should use default)",
			query: url.Values{
				"cursor": []string{"validCursor"},
				"search": []string{"query"},
			},
			expected: Paging{
				cursor: "validCursor",
				limit:  defaultReturnLimit,
				search: "query",
			},
		},
		{
			name:  "Missing all optional parameters (should use defaults)",
			query: url.Values{},
			expected: Paging{
				cursor: "",
				limit:  defaultReturnLimit,
				search: "",
			},
		},
		{
			name: "Invalid cursor (wrong format)",
			query: url.Values{
				"cursor": []string{"invalid@cursor"},
			},
			errMsg: "given value 'invalid@cursor' has incorrect format",
		},
		{
			name: "Invalid limit (negative value)",
			query: url.Values{
				"limit": []string{"-5"},
			},
			errMsg: "given value -5 is not a positive integer",
		},
		{
			name: "Limit above maximum",
			query: url.Values{
				"limit": []string{fmt.Sprintf("%d", maxQueryLimit+1)},
			},
			errMsg: fmt.Sprintf("given value '%d' is too large, maximum value is %d", maxQueryLimit+1, maxQueryLimit),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkPagingParams(tt.query)
			err := v.AsError()
			if tt.errMsg != "" {
				assert.Error(t, err, "Expected an error but got none")
				assert.ErrorContains(t, err, tt.errMsg)
			} else {
				assert.Nil(t, err, "Expected no errors but found some")
				assert.Equal(t, tt.expected, result, "Expected correct value")
			}
		})
	}
}

// Mock enum type for testing
type TestEnum int

const (
	TestEnumOne TestEnum = iota + 1
	TestEnumTwo
)

func (t TestEnum) Int() int {
	return int(t)
}

func (t TestEnum) NewFromString(s string) TestEnum {
	switch s {
	case "", "one":
		return TestEnumOne
	case "two", "2":
		return TestEnumTwo
	default:
		return -1
	}
}

// Implement EnumFactory interface for TestEnum
var _ enums.EnumFactory[TestEnum] = TestEnum(0)

func TestCheckEnum(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		expectErr bool
		expected  TestEnum
		errMsg    string
	}{
		{
			name:      "Valid enum value - one",
			param:     "one",
			expectErr: false,
			expected:  TestEnumOne,
		},
		{
			name:      "Valid enum value - empty string",
			param:     "",
			expectErr: false,
			expected:  TestEnumOne,
		},
		{
			name:      "Valid enum value - two",
			param:     "two",
			expectErr: false,
			expected:  TestEnumTwo,
		},
		{
			name:      "Valid enum value - two",
			param:     "2",
			expectErr: false,
			expected:  TestEnumTwo,
		},
		{
			name:      "Invalid enum value",
			param:     "invalid",
			expectErr: true,
			expected:  -1,
			errMsg:    "given value 'invalid' is not valid",
		},
		{
			name:      "Invalid enum value",
			param:     "One",
			expectErr: true,
			expected:  -1,
			errMsg:    "given value 'One' is not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := checkEnum[TestEnum](&v, tt.param, "enum_field")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.True(t, enums.IsInvalidEnum(result), "Expected invalid enum value")
				assert.Contains(t, v["enum_field"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
				assert.Equal(t, tt.expected, result, "Expected correct enum value")
			}
		})
	}
}

func TestParseSortOrder(t *testing.T) {
	tests := []validationTestCase[bool]{
		{
			name:     "Empty string (should return default)",
			param:    "",
			expected: defaultDesc,
		},
		{
			name:     "Ascending order",
			param:    "asc",
			expected: false,
		},
		{
			name:     "Descending order",
			param:    "desc",
			expected: true,
		},
		{
			name:     "Invalid sort order",
			param:    "random",
			expected: false, // Default return value in case of error
			errMsg:   "given value 'random' for parameter 'sort' is not valid, allowed order values are: asc, desc",
		},
		{
			name:     "Case-sensitive check (invalid ASC)",
			param:    "ASC",
			expected: false,
			errMsg:   "given value 'ASC' for parameter 'sort' is not valid, allowed order values are: asc, desc",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[bool]) bool {
		return v.parseSortOrder(tt.param)
	})
}

func TestCheckSort(t *testing.T) {
	tests := []validationTestCase[*types.Sort[TestEnum]]{
		{
			name:     "Valid sort with default order",
			param:    "one",
			expected: &types.Sort[TestEnum]{Column: TestEnumOne, Desc: defaultDesc},
		},
		{
			name:     "Valid sort with ascending order",
			param:    "one:asc",
			expected: &types.Sort[TestEnum]{Column: TestEnumOne, Desc: false},
		},
		{
			name:     "Valid sort with descending order",
			param:    "two:desc",
			expected: &types.Sort[TestEnum]{Column: TestEnumTwo, Desc: true},
		},
		{
			name:   "Invalid column name",
			param:  "invalid",
			errMsg: "given value 'invalid' is not valid",
		},
		{
			name:   "Invalid sort order",
			param:  "one:random",
			errMsg: "given value 'random' for parameter 'sort' is not valid, allowed order values are: asc, desc",
		},
		{
			name:   "Too many parts in sort string",
			param:  "one:desc:extra",
			errMsg: "given value 'one:desc:extra' for parameter 'sort' is not valid, expected format is '<column_name>[:(asc|desc)]'",
		},
		{
			name:     "Empty string (should return default enum and order)",
			param:    "",
			expected: &types.Sort[TestEnum]{Column: TestEnumOne, Desc: defaultDesc},
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[*types.Sort[TestEnum]]) *types.Sort[TestEnum] {
		return checkSort[TestEnum](v, tt.param)
	})
}

func TestCheckProtocolModes(t *testing.T) {
	tests := []validationTestCase[types.VDBProtocolModes]{
		{
			name:     "Valid single protocol mode",
			param:    "rocket_pool",
			expected: types.VDBProtocolModes{RocketPool: true},
		},
		{
			name:     "Valid protocol mode with spaces",
			param:    " rocket_pool ",
			expected: types.VDBProtocolModes{RocketPool: true},
		},
		{
			name:     "Empty string (should return empty struct)",
			param:    "",
			expected: types.VDBProtocolModes{},
		},
		{
			name:   "Invalid protocol mode",
			param:  "invalid_mode",
			errMsg: "given value 'invalid_mode' is not a valid protocol mode",
		},
		{
			name:     "Multiple valid protocol modes (should only enable rocket_pool)",
			param:    "rocket_pool,rocket_pool",
			expected: types.VDBProtocolModes{RocketPool: true},
		},
		{
			name:   "Valid and invalid protocol mode mixed",
			param:  "rocket_pool,invalid_mode",
			errMsg: "given value 'invalid_mode' is not a valid protocol mode",
		},
	}

	runValidationTests(t, tests, func(v *validationError, tt validationTestCase[types.VDBProtocolModes]) types.VDBProtocolModes {
		return v.checkProtocolModes(tt.param)
	})
}

func TestCheckValidatorList(t *testing.T) {
	tests := []struct {
		name         string
		param        string
		allowEmpty   bool
		expectedIdx  []types.VDBValidator
		expectedKeys []string
		errMsg       string
	}{
		{
			name:        "Valid single validator index",
			param:       "123",
			allowEmpty:  false,
			expectedIdx: []types.VDBValidator{123},
		},
		{
			name:        "Valid multiple validator indices",
			param:       "123,456,789",
			allowEmpty:  false,
			expectedIdx: []types.VDBValidator{123, 456, 789},
		},
		{
			name:         "Valid single public key",
			param:        "0x90ffa3d94a3b54f785087d9f9a0bb9925cb74cceb0759404b4ea85a7e46641eb45fd30e98c8bc289a8706733bf3a07dd",
			allowEmpty:   false,
			expectedKeys: []string{"0x90ffa3d94a3b54f785087d9f9a0bb9925cb74cceb0759404b4ea85a7e46641eb45fd30e98c8bc289a8706733bf3a07dd"},
		},
		{
			name:         "Valid multiple public keys",
			param:        "0x90ffa3d94a3b54f785087d9f9a0bb9925cb74cceb0759404b4ea85a7e46641eb45fd30e98c8bc289a8706733bf3a07dd,0xac701fb11446a7b0fe2dd2f10f07b6b899201cea9c102d5c5c3290fc05ef214645e4eb1466bb89ed0af4d2f16c901fc9",
			allowEmpty:   false,
			expectedKeys: []string{"0x90ffa3d94a3b54f785087d9f9a0bb9925cb74cceb0759404b4ea85a7e46641eb45fd30e98c8bc289a8706733bf3a07dd", "0xac701fb11446a7b0fe2dd2f10f07b6b899201cea9c102d5c5c3290fc05ef214645e4eb1466bb89ed0af4d2f16c901fc9"},
		},
		{
			name:        "Mixed indices and public keys",
			param:       "123,0x90ffa3d94a3b54f785087d9f9a0bb9925cb74cceb0759404b4ea85a7e46641eb45fd30e98c8bc289a8706733bf3a07dd,456",
			allowEmpty:  false,
			expectedIdx: []types.VDBValidator{123, 456},
			expectedKeys: []string{
				"0x90ffa3d94a3b54f785087d9f9a0bb9925cb74cceb0759404b4ea85a7e46641eb45fd30e98c8bc289a8706733bf3a07dd",
			},
		},
		{
			name:       "Empty string but allowEmpty = true",
			param:      "",
			allowEmpty: true,
		},
		{
			name:       "Empty string but allowEmpty = false",
			param:      "",
			allowEmpty: false,
			errMsg:     "list of validators must not be empty",
		},
		{
			name:       "Invalid index (non-numeric)",
			param:      "abc",
			allowEmpty: false,
			errMsg:     "invalid value",
		},
		{
			name:       "Invalid public key format",
			param:      "0xinvalidkey",
			allowEmpty: false,
			errMsg:     "invalid value",
		},
		{
			name:       "Hex decoding failure",
			param:      "0x123",
			allowEmpty: false,
			errMsg:     "invalid value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			indexes, publicKeys := v.checkValidatorList(tt.param, tt.allowEmpty)
			err := v.AsError()

			if tt.errMsg != "" {
				assert.Error(t, err, "Expected an error but got none")
				assert.ErrorContains(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err, "Expected no errors but found some: %w", v)
				assert.Equal(t, tt.expectedIdx, indexes, "Expected correct validator indices")
				assert.Equal(t, tt.expectedKeys, publicKeys, "Expected correct public keys")
			}
		})
	}
}

func TestCheckValidators(t *testing.T) {
	validInt1 := uint64(123)
	validInt2 := uint64(456)
	validKey1 := "0x90ffa3d94a3b54f785087d9f9a0bb9925cb74cceb0759404b4ea85a7e46641eb45fd30e98c8bc289a8706733bf3a07dd"
	validKey2 := "0xac701fb11446a7b0fe2dd2f10f07b6b899201cea9c102d5c5c3290fc05ef214645e4eb1466bb89ed0af4d2f16c901fc9"

	tests := []struct {
		name         string
		param        []intOrString
		allowEmpty   bool
		expectedIdx  []types.VDBValidator
		expectedKeys []string
		errMsg       string
	}{
		{
			name:       "Valid single validator index",
			param:      []intOrString{{intValue: &validInt1}},
			allowEmpty: false,
			expectedIdx: []types.VDBValidator{
				123,
			},
		},
		{
			name:       "Valid multiple validator indices",
			param:      []intOrString{{intValue: &validInt1}, {intValue: &validInt2}},
			allowEmpty: false,
			expectedIdx: []types.VDBValidator{
				123, 456,
			},
		},
		{
			name:         "Valid single public key",
			param:        []intOrString{{strValue: &validKey1}},
			allowEmpty:   false,
			expectedKeys: []string{validKey1},
		},
		{
			name:         "Valid multiple public keys",
			param:        []intOrString{{strValue: &validKey1}, {strValue: &validKey2}},
			allowEmpty:   false,
			expectedKeys: []string{validKey1, validKey2},
		},
		{
			name:       "Mixed indices and public keys",
			param:      []intOrString{{intValue: &validInt1}, {strValue: &validKey1}, {intValue: &validInt2}},
			allowEmpty: false,
			expectedIdx: []types.VDBValidator{
				123, 456,
			},
			expectedKeys: []string{validKey1},
		},
		{
			name:       "Empty list but allowEmpty = true",
			param:      []intOrString{},
			allowEmpty: true,
		},
		{
			name:       "Empty list but allowEmpty = false",
			param:      []intOrString{},
			allowEmpty: false,
			errMsg:     "list of validators is empty",
		},
		{
			name:       "Invalid public key format",
			param:      []intOrString{{strValue: new(string)}}, // Empty string as strValue
			allowEmpty: false,
			errMsg:     "given value '' is not a valid validator",
		},
		{
			name:       "Nil value in list",
			param:      []intOrString{{}},
			allowEmpty: false,
			errMsg:     "list contains invalid validator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			indexes, publicKeys := v.checkValidators(tt.param, tt.allowEmpty)
			err := v.AsError()
			if tt.errMsg != "" {
				assert.Error(t, err, "Expected an error but got none")
				assert.ErrorContains(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err, "Expected no errors but found some: %w", v)
				assert.Equal(t, tt.expectedIdx, indexes, "Expected correct validator indices")
				assert.Equal(t, tt.expectedKeys, publicKeys, "Expected correct public keys")
			}
		})
	}
}
