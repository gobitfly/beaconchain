package handlers

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"testing"

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
	assert.Nil(t, v)
	assert.False(t, v.hasErrors())
	v = validationError{}
	assert.NotNil(t, v)
	assert.False(t, v.hasErrors())
}

func TestValidationError_HasErrors_NonEmpty(t *testing.T) {
	var v validationError

	v.add("field1", "must be a valid email")

	assert.True(t, v.hasErrors())
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

			assert.Equal(t, tt.param, result, "Expected input to be returned unchanged")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["param"], fmt.Sprintf(`given value '%s' has incorrect format`, tt.param))
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
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid length",
			param:     "ValidName",
			minLength: 3,
			expectErr: false,
		},
		{
			name:      "Too short",
			param:     "Yo",
			minLength: 3,
			expectErr: true,
			errMsg:    "given value 'Yo' is too short, minimum length is 3",
		},
		{
			name:      "Too long",
			param:     strings.Repeat("a", maxNameLength*2),
			minLength: 3,
			expectErr: true,
			errMsg:    fmt.Sprintf("given value '%s' is too long, maximum length is %d", strings.Repeat("a", maxNameLength*2), maxNameLength),
		},
		{
			name:      "Exactly min length",
			param:     "Min",
			minLength: 3,
			expectErr: false,
		},
		{
			name:      "Exactly max length",
			param:     strings.Repeat("a", maxNameLength),
			minLength: 3,
			expectErr: false,
		},
		{
			name:      "Empty string",
			param:     "",
			minLength: 1,
			expectErr: true,
			errMsg:    "given value '' is too short, minimum length is 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkLength(tt.param, "param", tt.minLength)

			assert.Equal(t, tt.param, result, "Expected input to be returned unchanged")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["param"], tt.errMsg)
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
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid name",
			param:     "Valid_Name-123",
			minLength: 3,
			expectErr: false,
		},
		{
			name:      "Too short",
			param:     "Yo",
			minLength: 3,
			expectErr: true,
			errMsg:    "given value 'Yo' is too short, minimum length is 3",
		},
		{
			name:      "Invalid characters",
			param:     "Invalid@Name",
			minLength: 3,
			expectErr: true,
			errMsg:    "given value 'Invalid@Name' has incorrect format",
		},
		{
			name:      "Only spaces",
			param:     "   ",
			minLength: 3,
			expectErr: false, // Allowed by regex, could be restricted elsewhere
		},
		{
			name:      "Valid with space",
			param:     "John Doe",
			minLength: 3,
			expectErr: false,
		},
		{
			name:      "Empty string",
			param:     "",
			minLength: 1,
			expectErr: true,
			errMsg:    "given value '' is too short, minimum length is 1",
		},
		{
			name:      "Valid with dot",
			param:     "Dr. Smith",
			minLength: 3,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkName(tt.param, tt.minLength)

			assert.Equal(t, tt.param, result, "Expected input to be returned unchanged")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["name"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
			}
		})
	}
}

func TestCheckNameNotEmpty(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid name",
			param:     "John",
			expectErr: false,
		},
		{
			name:      "Empty string",
			param:     "",
			expectErr: true,
			errMsg:    "given value '' is too short, minimum length is 1",
		},
		{
			name:      "Only spaces",
			param:     "   ",
			expectErr: false, // Allowed by regex, could be restricted elsewhere
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkNameNotEmpty(tt.param)

			assert.Equal(t, tt.param, result, "Expected input to be returned unchanged")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["name"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
			}
		})
	}
}

func TestCheckKeyNotEmpty(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid key",
			param:     "ValidKey123",
			expectErr: false,
		},
		{
			name:      "Empty string",
			param:     "",
			expectErr: true,
			errMsg:    "given value '' is too short, minimum length is 1",
		},
		{
			name:      "Invalid characters",
			param:     "Invalid@Key",
			expectErr: true,
			errMsg:    "given value 'Invalid@Key' has incorrect format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkKeyNotEmpty(tt.param)

			assert.Equal(t, tt.param, result, "Expected input to be returned unchanged")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["key"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
			}
		})
	}
}

func TestCheckEmail(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid email",
			param:     "test@example.com",
			expectErr: false,
		},
		{
			name:      "Valid email with +",
			param:     "test+test@example.com",
			expectErr: false,
		},
		{
			name:      "Valid email with uppercase",
			param:     "Test@Example.COM",
			expectErr: false,
		},
		{
			name:      "Missing @ symbol",
			param:     "invalid-email.com",
			expectErr: true,
			errMsg:    "given value 'invalid-email.com' has incorrect format",
		},
		{
			name:      "Missing domain",
			param:     "user@",
			expectErr: true,
			errMsg:    "given value 'user@' has incorrect format",
		},
		{
			name:      "Empty string",
			param:     "",
			expectErr: true,
			errMsg:    "given value '' has incorrect format",
		},
		{
			name:      "Invalid special characters",
			param:     "user@exa mple.com",
			expectErr: true,
			errMsg:    "given value 'user@exa mple.com' has incorrect format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkEmail(tt.param)

			assert.Equal(t, strings.ToLower(tt.param), result, "Expected email to be returned in lowercase")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["email"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid password (minimum length)",
			param:     "12345",
			expectErr: false,
		},
		{
			name:      "Valid password (longer)",
			param:     "securePassword123!",
			expectErr: false,
		},
		{
			name:      "Too short",
			param:     "1234",
			expectErr: true,
			errMsg:    "given value '1234' has incorrect format",
		},
		{
			name:      "Empty string",
			param:     "",
			expectErr: true,
			errMsg:    "given value '' has incorrect format",
		},
		{
			name:      "Whitespace only (should still pass since regex only enforces length)",
			param:     "     ",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkPassword(tt.param)

			assert.Equal(t, tt.param, result, "Expected input to be returned unchanged")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["password"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
			}
		})
	}
}

func TestCheckUserEmailToken(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid token (40 lowercase alphanumeric characters)",
			param:     "abcd1234abcd1234abcd1234abcd1234abcd1234",
			expectErr: false,
		},
		{
			name:      "Too short",
			param:     "abcd1234abcd1234",
			expectErr: true,
			errMsg:    "given value 'abcd1234abcd1234' has incorrect format",
		},
		{
			name:      "Too long",
			param:     "abcd1234abcd1234abcd1234abcd1234abcd1234abcd",
			expectErr: true,
			errMsg:    "given value 'abcd1234abcd1234abcd1234abcd1234abcd1234abcd' has incorrect format",
		},
		{
			name:      "Contains uppercase letters",
			param:     "ABCD1234abcd1234abcd1234abcd1234abcd1234",
			expectErr: true,
			errMsg:    "given value 'ABCD1234abcd1234abcd1234abcd1234abcd1234' has incorrect format",
		},
		{
			name:      "Contains special characters",
			param:     "abcd1234abcd1234abcd1234abcd1234abcd12$%",
			expectErr: true,
			errMsg:    "given value 'abcd1234abcd1234abcd1234abcd1234abcd12$%' has incorrect format",
		},
		{
			name:      "Empty string",
			param:     "",
			expectErr: true,
			errMsg:    "given value '' has incorrect format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkUserEmailToken(tt.param)

			assert.Equal(t, tt.param, result, "Expected input to be returned unchanged")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["token"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
			}
		})
	}
}
func TestCheckValidatorDashboardPublicId(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		expectErr bool
		expected  types.VDBIdPublic
		errMsg    string
	}{
		{
			name:      "Valid public dashboard ID",
			param:     "v-12345678-1234-1234-1234-123456789abc",
			expectErr: false,
			expected:  types.VDBIdPublic("v-12345678-1234-1234-1234-123456789abc"),
		},
		{
			name:      "Invalid format (missing 'v-')",
			param:     "12345678-1234-1234-1234-123456789abc",
			expectErr: true,
			errMsg:    "given value '12345678-1234-1234-1234-123456789abc' has incorrect format",
		},
		{
			name:      "Invalid format (wrong character in UUID)",
			param:     "v-12345678-1234-1234-1234-123456789xyz",
			expectErr: true,
			errMsg:    "given value 'v-12345678-1234-1234-1234-123456789xyz' has incorrect format",
		},
		{
			name:      "Invalid format (too short)",
			param:     "v-12345678-1234-1234-1234-12345678",
			expectErr: true,
			errMsg:    "given value 'v-12345678-1234-1234-1234-12345678' has incorrect format",
		},
		{
			name:      "Empty string",
			param:     "",
			expectErr: true,
			errMsg:    "given value '' has incorrect format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkValidatorDashboardPublicId(tt.param)

			assert.Equal(t, types.VDBIdPublic(tt.param), result, "Expected input to be returned unchanged")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["public_dashboard_id"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
			}
		})
	}
}
func TestCheckAddress(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid address with 0x prefix",
			param:     "0x1234567890abcdef1234567890abcdef12345678",
			expectErr: false,
		},
		{
			name:      "Valid address without 0x prefix",
			param:     "1234567890abcdef1234567890abcdef12345678",
			expectErr: false,
		},
		{
			name:      "Invalid length (too short)",
			param:     "0x123456",
			expectErr: true,
			errMsg:    "given value '0x123456' has incorrect format",
		},
		{
			name:      "Invalid length (too long)",
			param:     "0x1234567890abcdef1234567890abcdef1234567890",
			expectErr: true,
			errMsg:    "given value '0x1234567890abcdef1234567890abcdef1234567890' has incorrect format",
		},
		{
			name:      "Invalid characters",
			param:     "0x1234567890abcdef1234567890abcdef1234567G",
			expectErr: true,
			errMsg:    "given value '0x1234567890abcdef1234567890abcdef1234567G' has incorrect format",
		},
		{
			name:      "Empty string",
			param:     "",
			expectErr: true,
			errMsg:    "given value '' has incorrect format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkAddress(tt.param)

			assert.Equal(t, tt.param, result, "Expected input to be returned unchanged")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["address"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
			}
		})
	}
}

func TestCheckInt(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		expectErr bool
		expected  int64
		errMsg    string
	}{
		{
			name:      "Valid positive integer",
			param:     "123",
			expectErr: false,
			expected:  123,
		},
		{
			name:      "Valid negative integer",
			param:     "-456",
			expectErr: false,
			expected:  -456,
		},
		{
			name:      "Zero",
			param:     "0",
			expectErr: false,
			expected:  0,
		},
		{
			name:      "Non-numeric string",
			param:     "abc",
			expectErr: true,
			errMsg:    "given value 'abc' is not an integer",
		},
		{
			name:      "Floating point number",
			param:     "3.14",
			expectErr: true,
			errMsg:    "given value '3.14' is not an integer",
		},
		{
			name:      "Empty string",
			param:     "",
			expectErr: true,
			errMsg:    "given value '' is not an integer",
		},
		{
			name:      "Large integer",
			param:     "9223372036854775807", // Max int64
			expectErr: false,
			expected:  math.MaxInt64,
		},
		{
			name:      "Too large for int64",
			param:     "9223372036854775808",
			expectErr: true,
			errMsg:    "given value '9223372036854775808' is not an integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkInt(tt.param, "param")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["param"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
				assert.Equal(t, tt.expected, result, "Expected integer conversion to be correct")
			}
		})
	}
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
	tests := []struct {
		name      string
		param     string
		expectErr bool
		expected  decimal.Decimal
		errMsg    string
	}{
		{
			name:      "Valid wei value",
			param:     "1000000000000000000",
			expectErr: false,
			expected:  decimal.RequireFromString("1000000000000000000"),
		},
		{
			name:      "Zero",
			param:     "0",
			expectErr: false,
			expected:  decimal.Zero,
		},
		{
			name:      "Negative number",
			param:     "-1",
			expectErr: true,
			errMsg:    "given value '-1' is not a wei string (must be positive integer)",
		},
		{
			name:      "Non-numeric string",
			param:     "abc",
			expectErr: true,
			errMsg:    "given value 'abc' is not a wei string (must be positive integer)",
		},
		{
			name:      "Floating point number",
			param:     "3.14",
			expectErr: true,
			errMsg:    "given value '3.14' is not a wei string (must be positive integer)",
		},
		{
			name:      "Empty string",
			param:     "",
			expectErr: true,
			errMsg:    "given value '' is not a wei string (must be positive integer)",
		},
		{
			name:      "Leading zeros (valid)",
			param:     "000123456789",
			expectErr: false,
			expected:  decimal.RequireFromString("123456789"),
		},
		{
			name:      "Very large wei value",
			param:     "115792089237316195423570985008687907853269984665640564039457584007913129639935", // max uint256
			expectErr: false,
			expected:  decimal.RequireFromString("115792089237316195423570985008687907853269984665640564039457584007913129639935"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkWeiDecimal(tt.param, "param")

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["param"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
				assert.True(t, tt.expected.Equal(result), "Expected decimal conversion to be correct")
			}
		})
	}
}

func TestCheckWeiMinMax(t *testing.T) {
	min := decimal.RequireFromString("1000")   // Minimum allowed value
	max := decimal.RequireFromString("100000") // Maximum allowed value

	tests := []struct {
		name      string
		param     string
		expectErr bool
		expected  decimal.Decimal
		errMsg    string
	}{
		{
			name:      "Valid value within range",
			param:     "50000",
			expectErr: false,
			expected:  decimal.RequireFromString("50000"),
		},
		{
			name:      "Exactly at minimum",
			param:     "1000",
			expectErr: false,
			expected:  decimal.RequireFromString("1000"),
		},
		{
			name:      "Exactly at maximum",
			param:     "100000",
			expectErr: false,
			expected:  decimal.RequireFromString("100000"),
		},
		{
			name:      "Below minimum",
			param:     "999",
			expectErr: true,
			errMsg:    "given value '999' is too small, minimum value is 1000",
		},
		{
			name:      "Above maximum",
			param:     "100001",
			expectErr: true,
			errMsg:    "given value '100001' is too large, maximum value is 100000",
		},
		{
			name:      "Invalid non-numeric input",
			param:     "abc",
			expectErr: true,
			errMsg:    "given value 'abc' is not a wei string (must be positive integer)",
		},
		{
			name:      "Empty string",
			param:     "",
			expectErr: true,
			errMsg:    "given value '' is not a wei string (must be positive integer)",
		},
		{
			name:      "Negative number",
			param:     "-5000",
			expectErr: true,
			errMsg:    "given value '-5000' is not a wei string (must be positive integer)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v validationError

			result := v.checkWeiMinMax(tt.param, "param", min, max)

			if tt.expectErr {
				assert.True(t, v.hasErrors(), "Expected an error but got none")
				assert.Contains(t, v["param"], tt.errMsg)
			} else {
				assert.False(t, v.hasErrors(), "Expected no errors but found some")
				assert.True(t, tt.expected.Equal(result), "Expected decimal conversion to be correct")
			}
		})
	}
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
