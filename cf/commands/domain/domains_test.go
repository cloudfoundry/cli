package domain_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/domain"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("domains command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
		domainRepo          *testapi.FakeDomainRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		domainRepo = &testapi.FakeDomainRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(NewListDomains(ui, configRepo, domainRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when an org is not targeted", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when not logged in", func() {
			requirementsFactory.TargetedOrgSuccess = true
			runCommand()
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails with usage when invoked with any args what so ever ", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true

			runCommand("whoops")
			Expect(ui.FailedWithUsage).To(BeTrue())
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
				domainRepo.ListDomainsForOrgDomains = []models.DomainFields{
					models.DomainFields{
						Shared: false,
						Name:   "Private-domain1",
					},
					models.DomainFields{
						Shared: true,
						Name:   "The-shared-domain",
					},
					models.DomainFields{
						Shared: false,
						Name:   "Private-domain2",
					},
				}
			})

			It("lists domains", func() {
				runCommand()

				Expect(domainRepo.ListDomainsForOrgGuid).To(Equal("my-org-guid"))
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
			domainRepo.ListDomainsForOrgApiResponse = errors.New("borked!")
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting domains in org", "my-org", "my-user"},
				[]string{"FAILED"},
				[]string{"Failed fetching domains"},
				[]string{"borked!"},
			))
		})
	})
})
