package helpers

import "strings"

// ConvertPathToRegularExpression converts a windows file path into a
// string which may be embedded in a ginkgo-compatible regular expression.
func ConvertPathToRegularExpression(path string) string {
	return strings.Replace(path, "\\", "\\\\", -1)
}
