package space_test

import (
	"github.com/cloudfoundry/cli/cf/api/space_quotas/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/space"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("space command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		quotaRepo           *fakes.FakeSpaceQuotaRepository
		configRepo          configuration.ReadWriter
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		quotaRepo = &fakes.FakeSpaceQuotaRepository{}
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(NewShowSpace(ui, configRepo, quotaRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.TargetedOrgSuccess = true
			runCommand("some-space")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when an org is not targeted", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("some-space")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when logged in and an org is targeted", func() {
		BeforeEach(func() {
			org := models.OrganizationFields{}
			org.Name = "my-org"

			app := models.ApplicationFields{}
			app.Name = "app1"
			app.Guid = "app1-guid"
			apps := []models.ApplicationFields{app}

			domain := models.DomainFields{}
			domain.Name = "domain1"
			domain.Guid = "domain1-guid"
			domains := []models.DomainFields{domain}

			serviceInstance := models.ServiceInstanceFields{}
			serviceInstance.Name = "service1"
			serviceInstance.Guid = "service1-guid"
			services := []models.ServiceInstanceFields{serviceInstance}

			securityGroup1 := models.SecurityGroupFields{Name: "Nacho Security"}
			securityGroup2 := models.SecurityGroupFields{Name: "Nacho Prime"}
			securityGroups := []models.SecurityGroupFields{securityGroup1, securityGroup2}

			space := models.Space{}
			space.Name = "whose-space-is-it-anyway"
			space.Organization = org
			space.Applications = apps
			space.Domains = domains
			space.ServiceInstances = services
			space.SecurityGroups = securityGroups
			space.SpaceQuotaGuid = "runaway-guid"

			quota := models.SpaceQuota{}
			quota.Guid = "runaway-guid"
			quota.Name = "runaway"
			quota.MemoryLimit = 102400
			quota.InstanceMemoryLimit = -1
			quota.RoutesLimit = 111
			quota.ServicesLimit = 222
			quota.NonBasicServicesAllowed = false

			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			requirementsFactory.Space = space

			quotaRepo.FindByGuidReturns(quota, nil)
		})

		Context("when the space has a space quota", func() {
			It("shows information about the given space", func() {
				runCommand("whose-space-is-it-anyway")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting info for space", "whose-space-is-it-anyway", "my-org", "my-user"},
					[]string{"OK"},
					[]string{"whose-space-is-it-anyway"},
					[]string{"Org", "my-org"},
					[]string{"Apps", "app1"},
					[]string{"Domains", "domain1"},
					[]string{"Services", "service1"},
					[]string{"Security Groups", "Nacho Security", "Nacho Prime"},
					[]string{"Space Quota", "runaway (100G memory limit, -1 instance memory limit, 111 routes, 222 services, paid services disallowed)"},
				))
			})
		})

		Context("when the space does not have a space quota", func() {
			It("shows information without a space quota", func() {
				requirementsFactory.Space.SpaceQuotaGuid = ""
				runCommand("whose-space-is-it-anyway")
				Expect(quotaRepo.FindByGuidCallCount()).To(Equal(0))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting info for space", "whose-space-is-it-anyway", "my-org", "my-user"},
					[]string{"OK"},
					[]string{"whose-space-is-it-anyway"},
					[]string{"Org", "my-org"},
					[]string{"Apps", "app1"},
					[]string{"Domains", "domain1"},
					[]string{"Services", "service1"},
					[]string{"Security Groups", "Nacho Security", "Nacho Prime"},
					[]string{"Space Quota"},
				))
			})
		})
	})
})
