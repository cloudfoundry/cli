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

var _ = Describe("Testing with ginkgo", func() {
	It("TestCreateDomainRequirements", func() {
		domainRepo := &testapi.FakeDomainRepository{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		callCreateDomain([]string{"my-org", "example.com"}, reqFactory, domainRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(reqFactory.OrganizationName).To(Equal("my-org"))

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}

		callCreateDomain([]string{"my-org", "example.com"}, reqFactory, domainRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestCreateDomainFailsWithUsage", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		domainRepo := &testapi.FakeDomainRepository{}
		ui := callCreateDomain([]string{""}, reqFactory, domainRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateDomain([]string{"org1"}, reqFactory, domainRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateDomain([]string{"org1", "example.com"}, reqFactory, domainRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestCreateDomain", func() {

		org := models.Organization{}
		org.Name = "myOrg"
		org.Guid = "myOrg-guid"
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}
		domainRepo := &testapi.FakeDomainRepository{}
		ui := callCreateDomain([]string{"myOrg", "example.com"}, reqFactory, domainRepo)

		Expect(domainRepo.CreateDomainName).To(Equal("example.com"))
		Expect(domainRepo.CreateDomainOwningOrgGuid).To(Equal("myOrg-guid"))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating domain", "example.com", "myOrg", "my-user"},
			{"OK"},
		})
	})
})

func callCreateDomain(args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-domain", args)

	token := configuration.TokenInfo{Username: "my-user"}
	configRepo := testconfig.NewRepositoryWithAccessToken(token)

	cmd := domain.NewCreateDomain(fakeUI, configRepo, domainRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
