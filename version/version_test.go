package version_test

import (
	"code.cloudfoundry.org/cli/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	Describe("VersionString", func() {
		When("passed no ldflags", func() {
			It("returns the default version", func() {
				Expect(version.VersionString()).To(Equal("0.0.0-unknown-version"))
			})
		})
	})
})
