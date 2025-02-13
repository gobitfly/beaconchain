package handlers

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/api/types"
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
