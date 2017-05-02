// +build !windows

package generic_test

import (
	"path/filepath"

	. "code.cloudfoundry.org/cli/util/generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExecutableFilename", func() {
	Context("when a filename which must be turned into an executable filename is input", func() {
		It("does nothing on unix", func() {
			myPath := filepath.Join("foo", "bar")
			Expect(ExecutableFilename(myPath)).To(Equal(myPath))
		})
	})
})
