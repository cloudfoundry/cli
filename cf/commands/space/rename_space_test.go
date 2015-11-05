package space_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("rename-space command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		spaceRepo           *testapi.FakeSpaceRepository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("rename-space").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		spaceRepo = &testapi.FakeSpaceRepository{}
	})

	var callRenameSpace = func(args []string) bool {
		return testcmd.RunCliCommand("rename-space", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("when the user is not logged in", func() {
		It("does not pass requirements", func() {
			requirementsFactory.LoginSuccess = false

			Expect(callRenameSpace([]string{"my-space", "my-new-space"})).To(BeFalse())
		})
	})

	Describe("when the user has not targeted an org", func() {
		It("does not pass requirements", func() {
			requirementsFactory.TargetedOrgSuccess = false

			Expect(callRenameSpace([]string{"my-space", "my-new-space"})).To(BeFalse())
		})
	})

	Describe("when the user provides fewer than two args", func() {
		It("fails with usage", func() {
			callRenameSpace([]string{"foo"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})
	})

	Describe("when the user is logged in and has provided an old and new space name", func() {
		BeforeEach(func() {
			space := models.Space{}
			space.Name = "the-old-space-name"
			space.Guid = "the-old-space-guid"
			requirementsFactory.Space = space
		})

		It("renames a space", func() {
			originalSpaceName := configRepo.SpaceFields().Name
			callRenameSpace([]string{"the-old-space-name", "my-new-space"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Renaming space", "the-old-space-name", "my-new-space", "my-org", "my-user"},
				[]string{"OK"},
			))

			spaceGUID, name := spaceRepo.RenameArgsForCall(0)
			Expect(spaceGUID).To(Equal("the-old-space-guid"))
			Expect(name).To(Equal("my-new-space"))
			Expect(configRepo.SpaceFields().Name).To(Equal(originalSpaceName))
		})

		Describe("renaming the space the user has targeted", func() {
			BeforeEach(func() {
				configRepo.SetSpaceFields(requirementsFactory.Space.SpaceFields)
			})

			It("renames the targeted space", func() {
				callRenameSpace([]string{"the-old-space-name", "my-new-space-name"})
				Expect(configRepo.SpaceFields().Name).To(Equal("my-new-space-name"))
			})
		})
	})
})
