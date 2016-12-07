package helpers

import (
	"os"
	"strconv"

	. "github.com/onsi/ginkgo"
)

// RunIfExperimental is for tests that should be skipped if CF_CLI_EXPERIMENTAL
// is set to false.
func RunIfExperimental(msg string) {
	if experimental, err := strconv.ParseBool(os.Getenv("CF_CLI_EXPERIMENTAL")); err != nil || !experimental {
		Skip("CF_CLI_EXPERIMENTAL=false - " + msg)
	}
}

// SkipIfExperimental is for tests that should be skipped if
// CF_CLI_EXPERIMENTAL is set to true.
func SkipIfExperimental(msg string) {
	if experimental, err := strconv.ParseBool(os.Getenv("CF_CLI_EXPERIMENTAL")); err == nil && experimental {
		Skip("CF_CLI_EXPERIMENTAL=true - " + msg)
	}
}
