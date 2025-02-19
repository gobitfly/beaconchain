package data

import (
	"reflect"
	"slices"
	"testing"
)

func TestReversePaddedIndex(t *testing.T) {
	test := []struct {
		name     string
		max      int
		inputs   []int
		expected []string
	}{
		{
			name:     "working",
			max:      100,
			inputs:   []int{1, 2},
			expected: []string{"99", "98"},
		},
		{
			name:   "bug example",
			max:    100,
			inputs: []int{0, 1, 90, 91, 100},
			// bug here, the result are not lexicographically reverse sorted
			// the resulting order is 91, 100, 90, 0, 1
			// the issues are with index == 0 and index == max
			expected: []string{"100", "99", "10", "09", "00"},
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			var res []string
			slices.Sort(tt.inputs)
			for _, input := range tt.inputs {
				res = append(res, reversePaddedIndex(input, tt.max))
			}
			if got, want := res, tt.expected; !reflect.DeepEqual(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}
