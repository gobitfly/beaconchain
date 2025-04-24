package databasetest

import (
	"strings"
	"testing"
)

func skipIfNoDocker(t testing.TB, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				if strings.Contains(err.Error(), "Docker not found") ||
					strings.Contains(err.Error(), "Is the docker daemon running?") {
					t.Skip("Docker not available")
				}
			}
			panic(r)
		}
	}()
	fn()
}
