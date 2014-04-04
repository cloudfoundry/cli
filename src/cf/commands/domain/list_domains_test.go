/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package domain_test

import (
	"cf/commands/domain"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("domains command", func() {
	It("TestListDomainsRequirements", func() {
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		domainRepo := &testapi.FakeDomainRepository{}

		callListDomains([]string{}, reqFactory, domainRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
		callListDomains([]string{}, reqFactory, domainRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callListDomains([]string{}, reqFactory, domainRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("TestListDomainsFailsWithUsage", func() {
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		domainRepo := &testapi.FakeDomainRepository{}

		ui := callListDomains([]string{"foo"}, reqFactory, domainRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("lists domains", func() {
		orgFields := models.OrganizationFields{}
		orgFields.Name = "my-org"
		orgFields.Guid = "my-org-guid"

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, OrganizationFields: orgFields}

		domainRepo := &testapi.FakeDomainRepository{
			ListDomainsForOrgDomains: []models.DomainFields{
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
				}},
		}

		ui := callListDomains([]string{}, reqFactory, domainRepo)

		Expect(domainRepo.ListDomainsForOrgGuid).To(Equal("my-org-guid"))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting domains in org", "my-org", "my-user"},
			{"name", "status"},
			{"The-shared-domain", "shared"},
			{"Private-domain1", "owned"},
			{"Private-domain2", "owned"},
		})
	})

	It("displays a message when no domains are found", func() {
		orgFields := models.OrganizationFields{}
		orgFields.Name = "my-org"
		orgFields.Guid = "my-org-guid"

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, OrganizationFields: orgFields}
		domainRepo := &testapi.FakeDomainRepository{}

		ui := callListDomains([]string{}, reqFactory, domainRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting domains in org", "my-org", "my-user"},
			{"No domains found"},
		})
	})

	It("fails when the domains API returns an error", func() {
		orgFields := models.OrganizationFields{}
		orgFields.Name = "my-org"
		orgFields.Guid = "my-org-guid"

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, OrganizationFields: orgFields}

		domainRepo := &testapi.FakeDomainRepository{
			ListDomainsForOrgApiResponse: errors.New("borked!"),
		}
		ui := callListDomains([]string{}, reqFactory, domainRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting domains in org", "my-org", "my-user"},
			{"failed"},
			{"Failed fetching domains"},
			{"borked!"},
		})
	})
})

func callListDomains(args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("domains", args)

	configRepo := testconfig.NewRepositoryWithAccessToken(configuration.TokenInfo{Username: "my-user"})

	spaceFields := models.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := models.OrganizationFields{}
	orgFields.Name = "my-org"

	configRepo.SetSpaceFields(spaceFields)
	configRepo.SetOrganizationFields(orgFields)

	cmd := domain.NewListDomains(fakeUI, configRepo, domainRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
