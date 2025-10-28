package version_test

import (
	"code.cloudfoundry.org/cli/v8/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	BeforeEach(func() {
		version.SetVersion("")
	})

	Describe("VersionString", func() {
		When("passed no ldflags", func() {
			It("returns the default version", func() {
				Expect(version.VersionString()).To(Equal("0.0.0-unknown-version"))
			})
		})

		When("a custom version is set", func() {
			It("returns the custom version", func() {
				version.SetVersion("1.2.3")
				Expect(version.VersionString()).To(Equal("1.2.3"))
			})
		})
	})

	Describe("SetVersion", func() {
		It("sets the version for valid semver versions", func() {
			version.SetVersion("1.2.3")
			Expect(version.VersionString()).To(Equal("1.2.3"))
		})

		It("exits with status code 1 when given an invalid semver", func() {
			var exitCode int
			originalExitFunc := version.ExitFunc

			defer func() {
				version.ExitFunc = originalExitFunc
			}()

			version.ExitFunc = func(code int) {
				exitCode = code
			}

			version.SetVersion("not-a-semver")
			Expect(exitCode).To(Equal(1))
		})
	})
})
