package testUtils

import (
	"os"
	"strings"
)

func HasTags(requiredTags ...string) bool {
	raw := os.Getenv("TAGS")
	if raw == "" {
		return false
	}

	set := make(map[string]bool)
	for _, tag := range strings.Split(raw, ",") {
		set[strings.TrimSpace(tag)] = true
	}

	for _, tag := range requiredTags {
		if set[tag] {
			return true
		}
	}
	return false
}
