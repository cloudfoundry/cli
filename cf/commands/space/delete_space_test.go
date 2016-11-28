package space_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
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

var _ = Describe("delete-space command", func() {
	var (
		ui                  *testterm.FakeUI
		space               models.Space
		config              coreconfig.Repository
		spaceRepo           *spacesfakes.FakeSpaceRepository
		orgRepo             *organizationsfakes.FakeOrganizationRepository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.Config = config
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("delete-space").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("delete-space", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		spaceRepo = new(spacesfakes.FakeSpaceRepository)
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		config = testconfig.NewRepositoryWithDefaults()

		space = models.Space{SpaceFields: models.SpaceFields{
			Name: "space-to-delete",
			GUID: "space-to-delete-guid",
		}}

		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
		spaceReq := new(requirementsfakes.FakeSpaceRequirement)
		spaceReq.GetSpaceReturns(space)
		requirementsFactory.NewSpaceRequirementReturns(spaceReq)
	})

	Describe("requirements", func() {
		BeforeEach(func() {
			ui.Inputs = []string{"y"}
		})
		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})

			Expect(runCommand("my-space")).To(BeFalse())
		})

		It("fails when not targeting an org and not providing -o", func() {
			targetedOrgReq := new(requirementsfakes.FakeTargetedOrgRequirement)
			targetedOrgReq.ExecuteReturns(errors.New("no org targeted"))
			requirementsFactory.NewTargetedOrgRequirementReturns(targetedOrgReq)

			Expect(runCommand("my-space")).To(BeFalse())
		})

		It("succeeds if you use the -o flag but don't have an org targeted", func() {
			targetedOrgReq := new(requirementsfakes.FakeTargetedOrgRequirement)
			targetedOrgReq.ExecuteReturns(errors.New("no org targeted"))
			requirementsFactory.NewTargetedOrgRequirementReturns(targetedOrgReq)

			Expect(runCommand("-o", "other-org", "my-space")).To(BeTrue())
		})
	})

	It("deletes a space, given its name", func() {
		ui.Inputs = []string{"yes"}
		runCommand("space-to-delete")

		Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the space space-to-delete"}))
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Deleting space", "space-to-delete", "my-org", "my-user"},
			[]string{"OK"},
		))
		Expect(spaceRepo.DeleteArgsForCall(0)).To(Equal("space-to-delete-guid"))
		Expect(config.HasSpace()).To(Equal(true))
	})

	It("does not prompt when the -f flag is given", func() {
		runCommand("-f", "space-to-delete")

		Expect(ui.Prompts).To(BeEmpty())
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Deleting", "space-to-delete"},
			[]string{"OK"},
		))
		Expect(spaceRepo.DeleteArgsForCall(0)).To(Equal("space-to-delete-guid"))
	})

	It("deletes a space in a different org, given the dash-o flag and a space-name", func() {
		otherSpace := models.Space{
			SpaceFields: models.SpaceFields{
				Name: "other-space-to-delete",
				GUID: "other-space-to-delete-guid",
			}}

		otherOrg := models.Organization{
			OrganizationFields: models.OrganizationFields{
				Name: "other-org",
				GUID: "other-org-guid",
			}}
		orgRepo.FindByNameReturns(otherOrg, nil)
		spaceRepo.FindByNameInOrgReturns(otherSpace, nil)

		ui.Inputs = []string{"yes"}
		runCommand("-o", "other-org", "other-space-to-delete")

		Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the space other-space-to-delete"}))
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Deleting space", "space-to-delete", "other-org", "my-user"},
			[]string{"OK"},
		))

		Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("other-org"))

		spaceArg, orgArg := spaceRepo.FindByNameInOrgArgsForCall(0)
		Expect(spaceArg).To(Equal("other-space-to-delete"))
		Expect(orgArg).To(Equal("other-org-guid"))
		Expect(spaceRepo.DeleteArgsForCall(0)).To(Equal("other-space-to-delete-guid"))
		Expect(config.HasSpace()).To(Equal(true))
	})

	It("clears the space from the config, when deleting the space currently targeted", func() {
		config.SetSpaceFields(space.SpaceFields)
		runCommand("-f", "space-to-delete")

		Expect(config.HasSpace()).To(Equal(false))
	})

	It("clears the space from the config, when deleting the space currently targeted even if space name is case insensitive", func() {
		config.SetSpaceFields(space.SpaceFields)
		runCommand("-f", "Space-To-Delete")

		Expect(config.HasSpace()).To(Equal(false))
	})
})
