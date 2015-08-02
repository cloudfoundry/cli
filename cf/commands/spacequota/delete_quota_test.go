package spacequota_test

import (
	test_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/api/space_quotas/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-quota command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *fakes.FakeSpaceQuotaRepository
		orgRepo             *test_org.FakeOrganizationRepository
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.Repository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetSpaceQuotaRepository(quotaRepo)
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("delete-space-quota").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		quotaRepo = &fakes.FakeSpaceQuotaRepository{}
		orgRepo = &test_org.FakeOrganizationRepository{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}

		org := models.Organization{}
		org.Name = "my-org"
		org.Guid = "my-org-guid"
		orgRepo.ListOrgsReturns([]models.Organization{org}, nil)
		orgRepo.FindByNameReturns(org, nil)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("delete-space-quota", args, requirementsFactory, updateCommandDependency, false)
	}

	Context("when the user is not logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = false
		})

		It("fails requirements", func() {
			Expect(runCommand("my-quota")).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("fails requirements when called without a quota name", func() {
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("fails requirements when an org is not targeted", func() {
			requirementsFactory.TargetedOrgSuccess = false
			Expect(runCommand()).To(BeFalse())
		})

		Context("When the quota provided exists", func() {
			BeforeEach(func() {
				quota := models.SpaceQuota{}
				quota.Name = "my-quota"
				quota.Guid = "my-quota-guid"
				quota.OrgGuid = "my-org-guid"
				quotaRepo.FindByNameReturns(quota, nil)
			})

			It("deletes a quota with a given name when the user confirms", func() {
				ui.Inputs = []string{"y"}

				runCommand("my-quota")
				Expect(quotaRepo.DeleteArgsForCall(0)).To(Equal("my-quota-guid"))

				Expect(ui.Prompts).To(ContainSubstrings(
					[]string{"Really delete the quota", "my-quota"},
				))

				Expect(ui.Outputs).To(ContainSubstrings(
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

				Expect(ui.Outputs).To(ContainSubstrings(
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

					Expect(ui.Outputs).To(ContainSubstrings(
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

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
					))

				})
			})
		})
	})
})
