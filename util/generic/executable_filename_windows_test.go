//go:build windows
// +build windows

package generic_test

import (
	"fmt"
	"path/filepath"

	. "code.cloudfoundry.org/cli/util/generic"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExecutableFilename", func() {
	When("a filename which must be turned into an executable filename is input", func() {
		It("appends .exe on Windows if it is not present", func() {
			myPath := filepath.Join("foo", "bar")
			Expect(ExecutableFilename(myPath)).To(Equal(fmt.Sprintf("%s.exe", myPath)))
		})

		It("doesn't append .exe on Windows if it is present", func() {
			myPath := filepath.Join("foo", "bar.exe")
			Expect(ExecutableFilename(myPath)).To(Equal(myPath))
		})
	})
})
