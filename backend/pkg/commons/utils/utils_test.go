package utils

import (
	"maps"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Define a generic test case type.
type sliceToMapTestCase[T comparable] struct {
	name     string
	input    []T
	expected map[T]struct{}
}

// runSliceToMapTests is a generic helper that executes all test cases for a given type T.
func runSliceToMapTests[T comparable](t *testing.T, testCases []sliceToMapTestCase[T]) {
	for _, tc := range testCases {
		// Capture tc to avoid issues with the loop variable in parallel tests.
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := SliceToMap(tc.input)
			assert.True(t, maps.Equal(tc.expected, result))
		})
	}
}

//
// Now, we can define our tests for various types.
//

// Test for int slices.
func TestSliceToMap_Int(t *testing.T) {
	testCases := []sliceToMapTestCase[int]{
		{
			name:     "nil slice",
			input:    nil,
			expected: map[int]struct{}{},
		},
		{
			name:     "empty slice",
			input:    []int{},
			expected: map[int]struct{}{},
		},
		{
			name:     "unique elements",
			input:    []int{1, 2, 3, 4, 5},
			expected: map[int]struct{}{1: {}, 2: {}, 3: {}, 4: {}, 5: {}},
		},
		{
			name:     "duplicate elements",
			input:    []int{1, 2, 2, 3, 3, 3, 4, 1, 1, 2, 2, 3, 3, 3, 1, 2, 4},
			expected: map[int]struct{}{1: {}, 2: {}, 3: {}, 4: {}},
		},
	}

	runSliceToMapTests(t, testCases)
}

// Test for string slices.
func TestSliceToMap_String(t *testing.T) {
	testCases := []sliceToMapTestCase[string]{
		{
			name:     "nil slice",
			input:    nil,
			expected: map[string]struct{}{},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: map[string]struct{}{},
		},
		{
			name:     "unique elements",
			input:    []string{"apple", "banana", "cherry"},
			expected: map[string]struct{}{"apple": {}, "banana": {}, "cherry": {}},
		},
		{
			name:     "duplicate elements",
			input:    []string{"apple", "banana", "apple", "cherry", "banana", "cherry"},
			expected: map[string]struct{}{"apple": {}, "banana": {}, "cherry": {}},
		},
	}

	runSliceToMapTests(t, testCases)
}

// A custom struct type that is comparable.
type testStruct struct {
	ID   int
	Name string
}

// Test for custom struct slices.
func TestSliceToMap_CustomType(t *testing.T) {
	a := testStruct{ID: 1, Name: "Alice"}
	b := testStruct{ID: 2, Name: "Bob"}
	c := testStruct{ID: 3, Name: "Charlie"}

	testCases := []sliceToMapTestCase[testStruct]{
		{
			name:     "nil slice",
			input:    nil,
			expected: map[testStruct]struct{}{},
		},
		{
			name:     "empty slice",
			input:    []testStruct{},
			expected: map[testStruct]struct{}{},
		},
		{
			name:     "unique elements",
			input:    []testStruct{a, b, c},
			expected: map[testStruct]struct{}{a: {}, b: {}, c: {}},
		},
		{
			name:     "duplicate elements",
			input:    []testStruct{a, b, a, a, b},
			expected: map[testStruct]struct{}{a: {}, b: {}},
		},
	}

	runSliceToMapTests(t, testCases)
}

func TestUint64Range(t *testing.T) {
	tests := []struct {
		name     string
		start    uint64
		end      uint64
		expected []uint64
	}{
		{
			name:     "Normal range",
			start:    1,
			end:      5,
			expected: []uint64{1, 2, 3, 4, 5},
		},
		{
			name:     "Single element range",
			start:    10,
			end:      10,
			expected: []uint64{10},
		},
		{
			name:     "Large range",
			start:    100,
			end:      105,
			expected: []uint64{100, 101, 102, 103, 104, 105},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Uint64Range(tt.start, tt.end)
			assert.Equal(t, tt.expected, result, "Uint64Range(%d, %d) = %v, expected %v", tt.start, tt.end, result, tt.expected)
		})
	}
}

func TestIterMap(t *testing.T) {
	t.Run("add 10 to each in int slice", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		want := []int{11, 12, 13, 14, 15}
		got := slices.Collect(
			IterMap(
				slices.Values(input),
				func(i int) int { return i + 10 },
			),
		)
		assert.Equal(t, want, got)
	})
	t.Run("convert int slice to string slice", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		want := []string{"1", "2", "3", "4", "5"}
		got := slices.Collect(
			IterMap(
				slices.Values(input),
				func(i int) string { return strconv.Itoa(i) },
			),
		)
		assert.Equal(t, want, got)
	})
}
