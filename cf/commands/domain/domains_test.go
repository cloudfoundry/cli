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
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetDomainRepository(domainRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("domains").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		domainRepo = &testapi.FakeDomainRepository{}
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
			BeforeEach(func() {
				domainRepo.ListDomainsForOrgStub = func(orgGuid string, cb func(models.DomainFields) bool) error {
					cb(models.DomainFields{
						Shared: false,
						Name:   "Private-domain1",
					})

					cb(models.DomainFields{
						Shared: true,
						Name:   "The-shared-domain",
					})

					cb(models.DomainFields{
						Shared: false,
						Name:   "Private-domain2",
					})

					return nil
				}
			})

			It("lists domains", func() {
				runCommand()

				orgGUID, _ := domainRepo.ListDomainsForOrgArgsForCall(0)
				Expect(orgGUID).To(Equal("my-org-guid"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting domains in org", "my-org", "my-user"},
					[]string{"name", "status"},
					[]string{"The-shared-domain", "shared"},
					[]string{"Private-domain1", "owned"},
					[]string{"Private-domain2", "owned"},
				))
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
			domainRepo.ListDomainsForOrgReturns(errors.New("an-error"))
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting domains in org", "my-org", "my-user"},
				[]string{"FAILED"},
				[]string{"Failed fetching domains"},
				[]string{"an-error"},
			))
		})
	})
})
