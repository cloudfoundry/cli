package requirements_test

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
	testassert "github.com/cloudfoundry/cli/testhelpers/assert"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("TargetedSpaceRequirement", func() {
	var (
		ui     *testterm.FakeUI
		config configuration.ReadWriter
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		config = testconfig.NewRepositoryWithDefaults()
	})

	Context("when the user has targeted a space", func() {
		It("succeeds", func() {
			req := NewTargetedSpaceRequirement(ui, config)
			Expect(req.Execute()).To(BeTrue())
		})
	})

	Context("when the user does not have a space targeted", func() {
		It("fails", func() {
			config.SetSpaceFields(models.SpaceFields{})

			testassert.AssertPanic(testterm.QuietPanic, func() {
				NewTargetedSpaceRequirement(ui, config).Execute()
			})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"No space targeted"},
			))
		})
	})
})
