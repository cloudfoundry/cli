package requirements_test

import (
	"code.cloudfoundry.org/cli/cf/requirements"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnsupportedLegacyFlagRequirement", func() {
	var (
		flags       []string
		requirement requirements.UnsupportedLegacyFlagRequirement
	)

	BeforeEach(func() {
		flags = []string{
			"--flag-1",
			"--flag-2",
			"--flag-3",
		}
		requirement = requirements.NewUnsupportedLegacyFlagRequirement(flags)
	})

	Context("Execute", func() {
		It("always returns an error", func() {
			Expect(requirement.Execute()).To(MatchError("The following flags cannot be used with deprecated usage: --flag-1, --flag-2, --flag-3"))
		})
	})
})
