package helpers

import (
	"fmt"
	"regexp"
	"runtime"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/types"
)

// SayPath is used to assert that a path is printed within streaming output.
// On Windows, it uses a case-insensitive match and escapes the path.
func SayPath(format string, path string) types.GomegaMatcher {
	if runtime.GOOS == "windows" {
		expected := "(?i)" + format
		expected = fmt.Sprintf(expected, regexp.QuoteMeta(path))
		return gbytes.Say(expected)
	}
	return gbytes.Say(format, path)
}
