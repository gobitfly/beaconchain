package benchmarks

import (
	"fmt"
	"strings"
	"testing"
)

func TestRandom(t *testing.T) {
	calc := 100000
	max := 1000000
	result := createRandomSeries(calc, max)
	fmt.Printf("%v", result)
	split := strings.Split(result, ",")
	if len(split) != calc {
		t.Errorf("Expected %d, got %d", calc, len(split))
	}
}
