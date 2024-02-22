package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Prompt asks for a string value using the label. For comand line interactions.
func CmdPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}
