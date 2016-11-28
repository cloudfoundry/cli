package space_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/spaces/spacesfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("rename-space command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		spaceRepo           *spacesfakes.FakeSpaceRepository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("rename-space").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		spaceRepo = new(spacesfakes.FakeSpaceRepository)
	})

	var callRenameSpace = func(args []string) bool {
		return testcmd.RunCLICommand("rename-space", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("when the user is not logged in", func() {
		It("does not pass requirements", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})

			Expect(callRenameSpace([]string{"my-space", "my-new-space"})).To(BeFalse())
		})
	})

	Describe("when the user has not targeted an org", func() {
		It("does not pass requirements", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			targetedOrgReq := new(requirementsfakes.FakeTargetedOrgRequirement)
			targetedOrgReq.ExecuteReturns(errors.New("no org targeted"))
			requirementsFactory.NewTargetedOrgRequirementReturns(targetedOrgReq)

			Expect(callRenameSpace([]string{"my-space", "my-new-space"})).To(BeFalse())
		})
	})

	Describe("when the user provides fewer than two args", func() {
		It("fails with usage", func() {
			callRenameSpace([]string{"foo"})
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})
	})

	Describe("when the user is logged in and has provided an old and new space name", func() {
		var space models.Space
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
			space = models.Space{}
			space.Name = "the-old-space-name"
			space.GUID = "the-old-space-guid"
			spaceReq := new(requirementsfakes.FakeSpaceRequirement)
			spaceReq.GetSpaceReturns(space)
			requirementsFactory.NewSpaceRequirementReturns(spaceReq)
		})

		It("renames a space", func() {
			originalSpaceName := configRepo.SpaceFields().Name
			callRenameSpace([]string{"the-old-space-name", "my-new-space"})

			Expect(ui.Outputs()).To(ContainSubstrings(
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
				configRepo.SetSpaceFields(space.SpaceFields)
			})

			It("renames the targeted space", func() {
				callRenameSpace([]string{"the-old-space-name", "my-new-space-name"})
				Expect(configRepo.SpaceFields().Name).To(Equal("my-new-space-name"))
			})
		})
	})
})
