package space_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/space"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func callSpaces(args []string, requirementsFactory *testreq.FakeReqFactory, config core_config.Reader, spaceRepo spaces.SpaceRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	cmd := NewListSpaces(ui, config, spaceRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}

var _ = Describe("spaces command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.ReadWriter
		spaceRepo           *testapi.FakeSpaceRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		spaceRepo = &testapi.FakeSpaceRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewListSpaces(ui, configRepo, spaceRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.TargetedOrgSuccess = true

			Expect(runCommand()).To(BeFalse())
		})

		It("fails when an org is not targeted", func() {
			requirementsFactory.LoginSuccess = true

			Expect(runCommand()).To(BeFalse())
		})
		It("should fail with usage when provided any arguments", func() {
			requirementsFactory.LoginSuccess = true
			Expect(runCommand("blahblah")).To(BeFalse())
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when logged in and an org is targeted", func() {
		BeforeEach(func() {
			space := models.Space{}
			space.Name = "space1"
			space2 := models.Space{}
			space2.Name = "space2"
			space3 := models.Space{}
			space3.Name = "space3"
			spaceRepo.Spaces = []models.Space{space, space2, space3}
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
		})

		It("lists all of the spaces", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting spaces in org", "my-org", "my-user"},
				[]string{"space1"},
				[]string{"space2"},
				[]string{"space3"},
			))
		})

		Context("when there are no spaces", func() {
			BeforeEach(func() {
				spaceRepo.Spaces = []models.Space{}
			})

			It("politely tells the user", func() {
				runCommand()
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting spaces in org", "my-org", "my-user"},
					[]string{"No spaces found"},
				))
			})
		})
	})
})
