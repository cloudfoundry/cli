package space_test

import (
	"github.com/cloudfoundry/cli/cf/api/space_quotas/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/space"
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

var _ = Describe("space command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		quotaRepo           *fakes.FakeSpaceQuotaRepository
		configRepo          core_config.ReadWriter
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		quotaRepo = &fakes.FakeSpaceQuotaRepository{}
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewShowSpace(ui, configRepo, quotaRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.TargetedOrgSuccess = true

			Expect(runCommand("some-space")).To(BeFalse())
		})

		It("fails when an org is not targeted", func() {
			requirementsFactory.LoginSuccess = true

			Expect(runCommand("some-space")).To(BeFalse())
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

			securityGroup1 := models.SecurityGroupFields{Name: "Nacho Security", Rules: []map[string]interface{}{
				{"protocol": "all", "destination": "0.0.0.0-9.255.255.255"},
			}}
			securityGroup2 := models.SecurityGroupFields{Name: "Nacho Prime", Rules: []map[string]interface{}{
				{"protocol": "udp", "ports": "8080-9090", "destination": "198.41.191.47/1"},
			}}
			securityGroups := []models.SecurityGroupFields{securityGroup1, securityGroup2}

			space := models.Space{}
			space.Name = "whose-space-is-it-anyway"
			space.Guid = "whose-space-is-it-anyway-guid"
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

		Context("when the guid flag is passed", func() {
			It("shows only the space guid", func() {
				runCommand("--guid", "whose-space-is-it-anyway")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"whose-space-is-it-anyway-guid"},
				))

				Expect(ui.Outputs).ToNot(ContainSubstrings(
					[]string{"Getting info for space", "whose-space-is-it-anyway", "my-org", "my-user"},
				))
			})
		})

		Context("when the security-group-rules flag is passed", func() {
			It("it shows space information and security group rules", func() {
				runCommand("--security-group-rules", "whose-space-is-it-anyway")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting rules for the security group", "Nacho Security"},
					[]string{"protocol", "all"},
					[]string{"destination", "0.0.0.0-9.255.255.255"},
					[]string{"Getting rules for the security group", "Nacho Prime"},
					[]string{"protocol", "udp"},
					[]string{"ports", "8080-9090"},
					[]string{"destination", "198.41.191.47/1"},
				))
			})
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
