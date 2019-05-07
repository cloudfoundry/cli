package helpers

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/types"
)

// SayPath is used to assert that a path is printed within streaming output.
// On Windows, it uses a case-insensitive match and escapes the path.
// On non-Windows, it evaluates the base directory of the path for symlinks.
func SayPath(format string, path string) types.GomegaMatcher {
	theRealDir, err := filepath.EvalSymlinks(filepath.Dir(path))
	Expect(err).ToNot(HaveOccurred())
	theRealPath := filepath.Join(theRealDir, filepath.Base(path))

	if runtime.GOOS == "windows" {
		expected := "(?i)" + format
		expected = fmt.Sprintf(expected, regexp.QuoteMeta(path))
		return gbytes.Say(expected)
	}
	return gbytes.Say(format, theRealPath)
}

func EqualPath(format string, path string) types.GomegaMatcher {
	theRealDir, err := filepath.EvalSymlinks(filepath.Dir(path))
	Expect(err).ToNot(HaveOccurred())
	theRealPath := filepath.Join(theRealDir, filepath.Base(path))

	if runtime.GOOS == "windows" {
		expected := "(?i)" + format
		expected = fmt.Sprintf(expected, regexp.QuoteMeta(path))
		return &matchers.MatchRegexpMatcher{
			Regexp: expected,
		}
	}

	return &matchers.MatchRegexpMatcher{
		Regexp: theRealPath,
	}
}
