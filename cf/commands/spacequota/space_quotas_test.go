package spacequota_test

import (
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

var _ = Describe("quotas command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *fakes.FakeSpaceQuotaRepository
		configRepo          core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetSpaceQuotaRepository(quotaRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("space-quotas").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		quotaRepo = &fakes.FakeSpaceQuotaRepository{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("space-quotas", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})

		It("requires the user to target an org", func() {
			requirementsFactory.TargetedOrgSuccess = false
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
		It("should fail with usage when provided any arguments", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			Expect(runCommand("blahblah")).To(BeFalse())
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "No argument"},
			))
		})
	})

	Context("when requirements have been met", func() {
		JustBeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			Expect(runCommand()).To(HavePassedRequirements())
		})

		Context("when quotas exist", func() {
			BeforeEach(func() {
				quotaRepo.FindByOrgReturns([]models.SpaceQuota{
					models.SpaceQuota{
						Name:                    "quota-name",
						MemoryLimit:             1024,
						InstanceMemoryLimit:     512,
						RoutesLimit:             111,
						ServicesLimit:           222,
						NonBasicServicesAllowed: true,
						OrgGuid:                 "my-org-guid",
					},
					models.SpaceQuota{
						Name:                    "quota-non-basic-not-allowed",
						MemoryLimit:             434,
						InstanceMemoryLimit:     -1,
						RoutesLimit:             1,
						ServicesLimit:           2,
						NonBasicServicesAllowed: false,
						OrgGuid:                 "my-org-guid",
					},
				}, nil)
			})

			It("lists quotas", func() {
				Expect(quotaRepo.FindByOrgArgsForCall(0)).To(Equal("my-org-guid"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting space quotas as", "my-user"},
					[]string{"OK"},
					[]string{"name", "total memory limit", "instance memory limit", "routes", "service instances", "paid service plans"},
					[]string{"quota-name", "1G", "512M", "111", "222", "allowed"},
					[]string{"quota-non-basic-not-allowed", "434M", "unlimited", "1", "2", "disallowed"},
				))
			})
			Context("when services are unlimited", func() {
				BeforeEach(func() {
					quotaRepo.FindByOrgReturns([]models.SpaceQuota{
						models.SpaceQuota{
							Name:                    "quota-non-basic-not-allowed",
							MemoryLimit:             434,
							InstanceMemoryLimit:     57,
							RoutesLimit:             1,
							ServicesLimit:           -1,
							NonBasicServicesAllowed: false,
							OrgGuid:                 "my-org-guid",
						},
					}, nil)
				})
				It("replaces -1 with unlimited", func() {
					Expect(quotaRepo.FindByOrgArgsForCall(0)).To(Equal("my-org-guid"))
					Expect(ui.Outputs).To(ContainSubstrings(

						[]string{"quota-non-basic-not-allowed", "434M", "57M ", "1", "unlimited", "disallowed"},
					))
				})

			})
		})

		Context("when an error occurs fetching quotas", func() {
			BeforeEach(func() {
				quotaRepo.FindByOrgReturns([]models.SpaceQuota{}, errors.New("I haz a borken!"))
			})

			It("prints an error", func() {
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting space quotas as", "my-user"},
					[]string{"FAILED"},
				))
			})
		})
	})

})
