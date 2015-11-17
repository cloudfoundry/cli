package domain_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("domains command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		domainRepo          *testapi.FakeDomainRepository
		routingApiRepo      *testapi.FakeRoutingApiRepository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetDomainRepository(domainRepo)
		deps.RepoLocator = deps.RepoLocator.SetRoutingApiRepository(routingApiRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("domains").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		domainRepo = &testapi.FakeDomainRepository{}
		routingApiRepo = &testapi.FakeRoutingApiRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("domains", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails when an org is not targeted", func() {
			requirementsFactory.LoginSuccess = true

			Expect(runCommand()).To(BeFalse())
		})

		It("fails when not logged in", func() {
			requirementsFactory.TargetedOrgSuccess = true

			Expect(runCommand()).To(BeFalse())
		})

		It("fails with usage when invoked with any args what so ever ", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true

			runCommand("whoops")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "No argument required"},
			))
		})
	})

	Context("when logged in and an org is targeted", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			requirementsFactory.OrganizationFields = models.OrganizationFields{
				Name: "my-org",
				Guid: "my-org-guid",
			}
		})

		Context("when there is at least one domain", func() {
			Context("when a domain has router group guid which is not found in routing api", func() {
				BeforeEach(func() {
					domainRepo.ListDomainsForOrgDomains = []models.DomainFields{
						models.DomainFields{
							Shared:          true,
							Name:            "The-shared-domain1",
							RouterGroupGuid: "valid_guid",
						},
						models.DomainFields{
							Shared:          true,
							Name:            "The-shared-domain2",
							RouterGroupGuid: "invalid_guid",
						},
					}

					routingApiRepo.RouterGroups = models.RouterGroups{
						models.RouterGroup{
							Guid: "valid_guid",
							Name: "default-router-group",
							Type: "tcp",
						},
					}
				})

				It("fails with error", func() {
					runCommand()

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Getting domains in org", "my-org", "my-user"},
						[]string{"FAILED"},
						[]string{"Invalid router group guid"},
					))
				})
			})

			Context("when all the domains' router group guids could be found", func() {
				BeforeEach(func() {
					domainRepo.ListDomainsForOrgDomains = []models.DomainFields{
						models.DomainFields{
							Shared: false,
							Name:   "Private-domain1",
						},
						models.DomainFields{
							Shared:          true,
							Name:            "The-shared-domain1",
							RouterGroupGuid: "tcp-router-group",
						},
						models.DomainFields{
							Shared:          true,
							Name:            "The-shared-domain2",
							RouterGroupGuid: "no-type-router-group",
						},
						models.DomainFields{
							Shared: false,
							Name:   "Private-domain2",
						},
					}

					routingApiRepo.RouterGroups = models.RouterGroups{
						models.RouterGroup{
							Guid: "tcp-router-group",
							Name: "default-router-group",
							Type: "tcp",
						},
						models.RouterGroup{
							Guid: "no-type-router-group",
							Name: "router-group-2",
							Type: "",
						},
					}
				})

				It("lists domains", func() {
					runCommand()

					Expect(domainRepo.ListDomainsForOrgGuid).To(Equal("my-org-guid"))
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Getting domains in org", "my-org", "my-user"},
						[]string{"name", "status", "routing"},
						[]string{"The-shared-domain1", "shared", "tcp"},
						[]string{"The-shared-domain2", "shared"},
						[]string{"Private-domain1", "owned"},
						[]string{"Private-domain2", "owned"},
					))
				})
			})
		})

		It("displays a message when no domains are found", func() {
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting domains in org", "my-org", "my-user"},
				[]string{"No domains found"},
			))
		})

		It("fails when the domains API returns an error", func() {
			domainRepo.ListDomainsForOrgApiResponse = errors.New("borked!")
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting domains in org", "my-org", "my-user"},
				[]string{"FAILED"},
				[]string{"Failed fetching domains"},
				[]string{"borked!"},
			))
		})

		It("fails with error when the routing api returns an error", func() {
			domainRepo.ListDomainsForOrgDomains = []models.DomainFields{
				models.DomainFields{
					Shared:          true,
					Name:            "Shared-domain",
					RouterGroupGuid: "router-group-guid",
				},
			}

			routingApiRepo.ListError = true
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting domains in org", "my-org", "my-user"},
				[]string{"FAILED"},
				[]string{"Failed fetching router groups"},
			))
		})
	})
})
