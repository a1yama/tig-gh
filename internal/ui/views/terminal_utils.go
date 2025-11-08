package views

import "strings"

func isTerminalResponse(key string) bool {
	return strings.Contains(key, ";rgb:") || strings.HasPrefix(key, "]")
}
