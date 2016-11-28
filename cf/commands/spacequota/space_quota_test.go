package spacequota_test

import (
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

var _ = Describe("quotas command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *spacequotasfakes.FakeSpaceQuotaRepository
		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetSpaceQuotaRepository(quotaRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("space-quota").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		quotaRepo = new(spacequotasfakes.FakeSpaceQuotaRepository)
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("space-quota", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("foo")).ToNot(HavePassedRequirements())
		})

		It("requires the user to target an org", func() {
			orgReq := new(requirementsfakes.FakeTargetedOrgRequirement)
			orgReq.ExecuteReturns(errors.New("not targeting org"))
			requirementsFactory.NewTargetedOrgRequirementReturns(orgReq)
			Expect(runCommand("bar")).ToNot(HavePassedRequirements())
		})

		It("fails when a quota name is not provided", func() {
			requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
	})

	Context("when logged in", func() {
		JustBeforeEach(func() {
			requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
			runCommand("quota-name")
		})

		Context("when quotas exist", func() {
			BeforeEach(func() {
				quotaRepo.FindByNameReturns(
					models.SpaceQuota{
						Name:                    "quota-name",
						MemoryLimit:             1024,
						InstanceMemoryLimit:     -1,
						RoutesLimit:             111,
						ServicesLimit:           222,
						NonBasicServicesAllowed: true,
						OrgGUID:                 "my-org-guid",
						AppInstanceLimit:        5,
						ReservedRoutePortsLimit: "4",
					}, nil)
			})

			It("lists the specific quota info", func() {
				Expect(quotaRepo.FindByNameArgsForCall(0)).To(Equal("quota-name"))
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting space quota quota-name info as", "my-user"},
					[]string{"OK"},
					[]string{"total memory limit", "1G"},
					[]string{"instance memory limit", "unlimited"},
					[]string{"routes", "111"},
					[]string{"service", "222"},
					[]string{"non basic services", "allowed"},
					[]string{"app instance limit", "5"},
					[]string{"reserved route ports", "4"},
				))
			})

			Context("when the services are unlimited", func() {
				BeforeEach(func() {
					quotaRepo.FindByNameReturns(
						models.SpaceQuota{
							Name:                    "quota-name",
							MemoryLimit:             1024,
							InstanceMemoryLimit:     14,
							RoutesLimit:             111,
							ServicesLimit:           -1,
							NonBasicServicesAllowed: true,
							OrgGUID:                 "my-org-guid",
						}, nil)

				})

				It("replaces -1 with unlimited", func() {
					Expect(quotaRepo.FindByNameArgsForCall(0)).To(Equal("quota-name"))
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Getting space quota quota-name info as", "my-user"},
						[]string{"OK"},
						[]string{"total memory limit", "1G"},
						[]string{"instance memory limit", "14M"},
						[]string{"routes", "111"},
						[]string{"service", "unlimited"},
						[]string{"non basic services", "allowed"},
					))
				})
			})

			Context("when the app instances are unlimited", func() {
				BeforeEach(func() {
					quotaRepo.FindByNameReturns(
						models.SpaceQuota{
							Name:                    "quota-name",
							MemoryLimit:             1024,
							InstanceMemoryLimit:     -1,
							RoutesLimit:             111,
							ServicesLimit:           222,
							NonBasicServicesAllowed: true,
							OrgGUID:                 "my-org-guid",
							AppInstanceLimit:        -1,
						}, nil)
				})

				It("replaces -1 with unlimited", func() {
					Expect(quotaRepo.FindByNameArgsForCall(0)).To(Equal("quota-name"))
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Getting space quota quota-name info as", "my-user"},
						[]string{"OK"},
						[]string{"total memory limit", "1G"},
						[]string{"instance memory limit", "unlimited"},
						[]string{"routes", "111"},
						[]string{"service", "222"},
						[]string{"non basic services", "allowed"},
						[]string{"app instance limit", "unlimited"},
					))
				})
			})
		})
		Context("when an error occurs fetching quotas", func() {
			BeforeEach(func() {
				quotaRepo.FindByNameReturns(models.SpaceQuota{}, errors.New("I haz a borken!"))
			})

			It("prints an error", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting space quota quota-name info as", "my-user"},
					[]string{"FAILED"},
				))
			})
		})
	})

})
