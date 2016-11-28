package organization_test

import (
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-org command", func() {
	var (
		config              coreconfig.Repository
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		orgRepo             *organizationsfakes.FakeOrganizationRepository
		org                 models.Organization
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.Config = config
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("delete-org").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{
			Inputs: []string{"y"},
		}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)

		org = models.Organization{}
		org.Name = "org-to-delete"
		org.GUID = "org-to-delete-guid"
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)

		orgRepo.ListOrgsReturns([]models.Organization{org}, nil)
		orgRepo.FindByNameReturns(org, nil)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("delete-org", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	It("fails requirements when not logged in", func() {
		requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
		Expect(runCommand("some-org-name")).To(BeFalse())
	})

	It("fails with usage if no arguments are given", func() {
		runCommand()
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires an argument"},
		))
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		})

		Context("when deleting the currently targeted org", func() {
			It("untargets the deleted org", func() {
				config.SetOrganizationFields(org.OrganizationFields)
				runCommand("org-to-delete")

				Expect(config.OrganizationFields()).To(Equal(models.OrganizationFields{}))
				Expect(config.SpaceFields()).To(Equal(models.SpaceFields{}))
			})
		})

		Context("when deleting an org that is not targeted", func() {
			BeforeEach(func() {
				otherOrgFields := models.OrganizationFields{}
				otherOrgFields.GUID = "some-other-org-guid"
				otherOrgFields.Name = "some-other-org"
				config.SetOrganizationFields(otherOrgFields)

				spaceFields := models.SpaceFields{}
				spaceFields.Name = "some-other-space"
				config.SetSpaceFields(spaceFields)
			})

			It("deletes the org with the given name", func() {
				runCommand("org-to-delete")

				Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the org org-to-delete"}))

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting", "org-to-delete"},
					[]string{"OK"},
				))

				Expect(orgRepo.DeleteArgsForCall(0)).To(Equal("org-to-delete-guid"))
			})

			It("does not untarget the org and space", func() {
				runCommand("org-to-delete")

				Expect(config.OrganizationFields().Name).To(Equal("some-other-org"))
				Expect(config.SpaceFields().Name).To(Equal("some-other-space"))
			})
		})

		It("does not prompt when the -f flag is given", func() {
			ui.Inputs = []string{}
			runCommand("-f", "org-to-delete")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Deleting", "org-to-delete"},
				[]string{"OK"},
			))

			Expect(orgRepo.DeleteArgsForCall(0)).To(Equal("org-to-delete-guid"))
		})

		It("warns the user when the org does not exist", func() {
			orgRepo.FindByNameReturns(models.Organization{}, errors.NewModelNotFoundError("Organization", "org org-to-delete does not exist"))

			runCommand("org-to-delete")

			Expect(orgRepo.DeleteCallCount()).To(Equal(0))

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Deleting", "org-to-delete"},
				[]string{"OK"},
			))
			Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"org-to-delete", "does not exist."}))
		})
	})
})
