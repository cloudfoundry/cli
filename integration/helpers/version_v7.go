// +build V7 V8

package helpers

import (
	"fmt"

	. "github.com/onsi/ginkgo"
)

const V7 = true

// SkipIfV7AndVersionLessThan is used to skip tests if the target build is V7 and API version < the specified version
// If minVersion contains the prefix 3 then the v3 version is checked, otherwise the v2 version is used.
func SkipIfV7AndVersionLessThan(minVersion string) {
	SkipIfVersionLessThan(minVersion)
}

// SkipIfV7 is used to skip tests if the target build is V7.
func SkipIfV7() {
	Skip(fmt.Sprintf("Not implemented for V7 yet"))
}
