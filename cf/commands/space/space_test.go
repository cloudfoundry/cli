package space_test

import (
	"github.com/cloudfoundry/cli/cf/api/space_quotas/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/plugin/models"
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
		configRepo          core_config.Repository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetSpaceQuotaRepository(quotaRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("space").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		quotaRepo = &fakes.FakeSpaceQuotaRepository{}
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}

		deps = command_registry.NewDependency()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("space", args, requirementsFactory, updateCommandDependency, false)
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

		It("Shows usage when called incorrectly", func() {
			requirementsFactory.LoginSuccess = true

			runCommand("some-space", "much")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})
	})

	Context("when logged in and an org is targeted", func() {
		BeforeEach(func() {
			org := models.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

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
				{"protocol": "all", "destination": "0.0.0.0-9.255.255.255", "log": true, "IntTest": 1000},
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
					[]string{"log", "true"},
					[]string{"IntTest", "1000"},
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

		Context("When called as a plugin", func() {
			var (
				pluginModel plugin_models.GetSpace_Model
			)
			BeforeEach(func() {
				pluginModel = plugin_models.GetSpace_Model{}
				deps.PluginModels.Space = &pluginModel
			})

			It("Fills in the PluginModel", func() {
				testcmd.RunCliCommand("space", []string{"whose-space-is-it-anyway"}, requirementsFactory, updateCommandDependency, true)
				Ω(pluginModel.Name).To(Equal("whose-space-is-it-anyway"))
				Ω(pluginModel.Guid).To(Equal("whose-space-is-it-anyway-guid"))

				Ω(pluginModel.Organization.Name).To(Equal("my-org"))
				Ω(pluginModel.Organization.Guid).To(Equal("my-org-guid"))

				Ω(pluginModel.Applications).To(HaveLen(1))
				Ω(pluginModel.Applications[0].Name).To(Equal("app1"))
				Ω(pluginModel.Applications[0].Guid).To(Equal("app1-guid"))

				Ω(pluginModel.Domains).To(HaveLen(1))
				Ω(pluginModel.Domains[0].Name).To(Equal("domain1"))
				Ω(pluginModel.Domains[0].Guid).To(Equal("domain1-guid"))

				Ω(pluginModel.ServiceInstances).To(HaveLen(1))
				Ω(pluginModel.ServiceInstances[0].Name).To(Equal("service1"))
				Ω(pluginModel.ServiceInstances[0].Guid).To(Equal("service1-guid"))

				Ω(pluginModel.SecurityGroups).To(HaveLen(2))
				Ω(pluginModel.SecurityGroups[0].Name).To(Equal("Nacho Security"))
				Ω(pluginModel.SecurityGroups[0].Rules).To(HaveLen(1))
				Ω(pluginModel.SecurityGroups[0].Rules[0]).To(HaveLen(4))
				val := pluginModel.SecurityGroups[0].Rules[0]["protocol"]
				Ω(val).To(Equal("all"))
				val = pluginModel.SecurityGroups[0].Rules[0]["destination"]
				Ω(val).To(Equal("0.0.0.0-9.255.255.255"))

				Ω(pluginModel.SecurityGroups[1].Name).To(Equal("Nacho Prime"))
				Ω(pluginModel.SecurityGroups[1].Rules).To(HaveLen(1))
				Ω(pluginModel.SecurityGroups[1].Rules[0]).To(HaveLen(3))
				val = pluginModel.SecurityGroups[1].Rules[0]["protocol"]
				Ω(val).To(Equal("udp"))
				val = pluginModel.SecurityGroups[1].Rules[0]["destination"]
				Ω(val).To(Equal("198.41.191.47/1"))
				val = pluginModel.SecurityGroups[1].Rules[0]["ports"]
				Ω(val).To(Equal("8080-9090"))

				Ω(pluginModel.SpaceQuota.Name).To(Equal("runaway"))
				Ω(pluginModel.SpaceQuota.Guid).To(Equal("runaway-guid"))
				Ω(pluginModel.SpaceQuota.MemoryLimit).To(Equal(int64(102400)))
				Ω(pluginModel.SpaceQuota.InstanceMemoryLimit).To(Equal(int64(-1)))
				Ω(pluginModel.SpaceQuota.RoutesLimit).To(Equal(111))
				Ω(pluginModel.SpaceQuota.ServicesLimit).To(Equal(222))
				Ω(pluginModel.SpaceQuota.NonBasicServicesAllowed).To(BeFalse())
			})
		})
	})

})
