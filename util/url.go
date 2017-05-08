package util

import "strings"

func IsHTTPScheme(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

func IsUnsupportedURLScheme(path string) bool {
	return strings.Contains(path, "://") && !IsHTTPScheme(path)
}
