package quota_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cli/cf/api/quotas/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("quota", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		config              core_config.Repository
		quotaRepo           *fakes.FakeQuotaRepository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetQuotaRepository(quotaRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("quota").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		quotaRepo = &fakes.FakeQuotaRepository{}
		config = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("quota", args, requirementsFactory, updateCommandDependency, false)
	}

	Context("When not logged in", func() {
		It("fails requirements", func() {
			Expect(runCommand("quota-name")).To(BeFalse())
		})
	})

	Context("When logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("When not providing a quota name", func() {
			It("fails with usage", func() {
				runCommand()
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Incorrect Usage", "Requires", "argument"},
				))
			})
		})

		Context("When providing a quota name", func() {
			Context("that exists", func() {
				BeforeEach(func() {
					quotaRepo.FindByNameReturns(models.QuotaFields{
						Guid:                    "my-quota-guid",
						Name:                    "muh-muh-muh-my-qua-quota",
						MemoryLimit:             512,
						InstanceMemoryLimit:     5,
						RoutesLimit:             2000,
						ServicesLimit:           47,
						NonBasicServicesAllowed: true,
					}, nil)
				})

				It("shows you that quota", func() {
					runCommand("muh-muh-muh-my-qua-quota")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Getting quota", "muh-muh-muh-my-qua-quota", "my-user"},
						[]string{"OK"},
						[]string{"Total Memory", "512M"},
						[]string{"Instance Memory", "5M"},
						[]string{"Routes", "2000"},
						[]string{"Services", "47"},
						[]string{"Paid service plans", "allowed"},
					))
				})
			})

			Context("when instance memory limit is -1", func() {
				BeforeEach(func() {
					quotaRepo.FindByNameReturns(models.QuotaFields{
						Guid:                    "my-quota-guid",
						Name:                    "muh-muh-muh-my-qua-quota",
						MemoryLimit:             512,
						InstanceMemoryLimit:     -1,
						RoutesLimit:             2000,
						ServicesLimit:           47,
						NonBasicServicesAllowed: true,
					}, nil)
				})

				It("shows you that quota", func() {
					runCommand("muh-muh-muh-my-qua-quota")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Getting quota", "muh-muh-muh-my-qua-quota", "my-user"},
						[]string{"OK"},
						[]string{"Total Memory", "512M"},
						[]string{"Instance Memory", "unlimited"},
						[]string{"Routes", "2000"},
						[]string{"Services", "47"},
						[]string{"Paid service plans", "allowed"},
					))
				})
			})

			Context("when the services limit is -1", func() {
				BeforeEach(func() {
					quotaRepo.FindByNameReturns(models.QuotaFields{
						Guid:                    "my-quota-guid",
						Name:                    "muh-muh-muh-my-qua-quota",
						MemoryLimit:             512,
						InstanceMemoryLimit:     14,
						RoutesLimit:             2000,
						ServicesLimit:           -1,
						NonBasicServicesAllowed: true,
					}, nil)
				})

				It("shows you that quota", func() {
					runCommand("muh-muh-muh-my-qua-quota")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Getting quota", "muh-muh-muh-my-qua-quota", "my-user"},
						[]string{"OK"},
						[]string{"Total Memory", "512M"},
						[]string{"Instance Memory", "14M"},
						[]string{"Routes", "2000"},
						[]string{"Services", "unlimited"},
						[]string{"Paid service plans", "allowed"},
					))
				})
			})

			Context("that doesn't exist", func() {
				BeforeEach(func() {
					quotaRepo.FindByNameReturns(models.QuotaFields{}, errors.New("oops i accidentally a quota"))
				})

				It("gives an error", func() {
					runCommand("an-quota")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"oops"},
					))
				})
			})
		})
	})
})
