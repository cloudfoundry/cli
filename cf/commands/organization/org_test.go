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

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package organization_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/organization"
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

var _ = Describe("Testing with ginkgo", func() {
	It("TestShowOrgRequirements", func() {
		args := []string{"my-org"}

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		callShowOrg(args, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}
		callShowOrg(args, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestShowOrgFailsWithUsage", func() {

		org := models.Organization{}
		org.Name = "my-org"
		org.Guid = "my-org-guid"
		requirementsFactory := &testreq.FakeReqFactory{Organization: org, LoginSuccess: true}

		args := []string{"my-org"}
		ui := callShowOrg(args, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeFalse())

		args = []string{}
		ui = callShowOrg(args, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})
	It("TestRunWhenOrganizationExists", func() {

		developmentSpaceFields := models.SpaceFields{}
		developmentSpaceFields.Name = "development"
		stagingSpaceFields := models.SpaceFields{}
		stagingSpaceFields.Name = "staging"
		domainFields := models.DomainFields{}
		domainFields.Name = "cfapps.io"
		cfAppDomainFields := models.DomainFields{}
		cfAppDomainFields.Name = "cf-app.com"
		org := models.Organization{}
		org.Name = "my-org"
		org.Guid = "my-org-guid"
		org.QuotaDefinition = models.NewQuotaFields("cantina-quota", 512, 2, 5, true)
		org.Spaces = []models.SpaceFields{developmentSpaceFields, stagingSpaceFields}
		org.Domains = []models.DomainFields{domainFields, cfAppDomainFields}

		requirementsFactory := &testreq.FakeReqFactory{Organization: org, LoginSuccess: true}

		args := []string{"my-org"}
		ui := callShowOrg(args, requirementsFactory)

		Expect(requirementsFactory.OrganizationName).To(Equal("my-org"))

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting info for org", "my-org", "my-user"},
			[]string{"OK"},
			[]string{"my-org"},
			[]string{"  domains:", "cfapps.io", "cf-app.com"},
			[]string{"  quota: ", "cantina-quota", "512M", "2 routes", "5 services", "paid services allowed"},
			[]string{"  spaces:", "development", "staging"},
		))
	})
})

func callShowOrg(args []string, requirementsFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	token := configuration.TokenInfo{Username: "my-user"}

	spaceFields := models.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := models.OrganizationFields{}
	orgFields.Name = "my-org"

	configRepo := testconfig.NewRepositoryWithAccessToken(token)
	configRepo.SetSpaceFields(spaceFields)
	configRepo.SetOrganizationFields(orgFields)

	cmd := NewShowOrg(ui, configRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}
