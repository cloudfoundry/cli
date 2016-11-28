package spacequota_test

import (
	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/api/spacequotas/spacequotasfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-space-quota command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *spacequotasfakes.FakeSpaceQuotaRepository
		orgRepo             *organizationsfakes.FakeOrganizationRepository
		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetSpaceQuotaRepository(quotaRepo)
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("delete-space-quota").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		quotaRepo = new(spacequotasfakes.FakeSpaceQuotaRepository)
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)

		org := models.Organization{}
		org.Name = "my-org"
		org.GUID = "my-org-guid"
		orgRepo.ListOrgsReturns([]models.Organization{org}, nil)
		orgRepo.FindByNameReturns(org, nil)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("delete-space-quota", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Context("when the user is not logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
		})

		It("fails requirements", func() {
			Expect(runCommand("my-quota")).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		})

		It("fails requirements when called without a quota name", func() {
			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("fails requirements when an org is not targeted", func() {
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})
			Expect(runCommand()).To(BeFalse())
		})

		Context("When the quota provided exists", func() {
			BeforeEach(func() {
				quota := models.SpaceQuota{}
				quota.Name = "my-quota"
				quota.GUID = "my-quota-guid"
				quota.OrgGUID = "my-org-guid"
				quotaRepo.FindByNameReturns(quota, nil)
			})

			It("deletes a quota with a given name when the user confirms", func() {
				ui.Inputs = []string{"y"}

				runCommand("my-quota")
				Expect(quotaRepo.DeleteArgsForCall(0)).To(Equal("my-quota-guid"))

				Expect(ui.Prompts).To(ContainSubstrings(
					[]string{"Really delete the quota", "my-quota"},
				))

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting space quota", "my-quota", "as", "my-user"},
					[]string{"OK"},
				))
			})

			It("does not prompt when the -f flag is provided", func() {
				runCommand("-f", "my-quota")

				Expect(quotaRepo.DeleteArgsForCall(0)).To(Equal("my-quota-guid"))

				Expect(ui.Prompts).To(BeEmpty())
			})

			It("shows an error when deletion fails", func() {
				quotaRepo.DeleteReturns(errors.New("some error"))

				runCommand("-f", "my-quota")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting", "my-quota"},
					[]string{"FAILED"},
				))
			})
		})

		Context("when finding the quota fails", func() {
			Context("when the quota provided does not exist", func() {
				BeforeEach(func() {
					quotaRepo.FindByNameReturns(models.SpaceQuota{}, errors.NewModelNotFoundError("Quota", "non-existent-quota"))
				})

				It("warns the user when that the quota does not exist", func() {
					runCommand("-f", "non-existent-quota")

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Deleting", "non-existent-quota"},
						[]string{"OK"},
					))

					Expect(ui.WarnOutputs).To(ContainSubstrings(
						[]string{"non-existent-quota", "does not exist"},
					))
				})
			})

			Context("when other types of error occur", func() {
				BeforeEach(func() {
					quotaRepo.FindByNameReturns(models.SpaceQuota{}, errors.New("some error"))
				})

				It("shows an error", func() {
					runCommand("-f", "my-quota")

					Expect(ui.WarnOutputs).ToNot(ContainSubstrings(
						[]string{"my-quota", "does not exist"},
					))

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
					))

				})
			})
		})
	})
})
