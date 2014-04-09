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

package organization_test

import (
	"cf/commands/organization"
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

func callListQuotas(requirementsFactory *testreq.FakeReqFactory, quotaRepo *testapi.FakeQuotaRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("quotas", []string{})

	spaceFields := models.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := models.OrganizationFields{}
	orgFields.Name = "my-org"

	token := configuration.TokenInfo{Username: "my-user"}
	config := testconfig.NewRepositoryWithAccessToken(token)
	config.SetSpaceFields(spaceFields)
	config.SetOrganizationFields(orgFields)

	cmd := organization.NewListQuotas(fakeUI, config, quotaRepo)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestListQuotasRequirements", func() {
		quotaRepo := &testapi.FakeQuotaRepository{}

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		callListQuotas(requirementsFactory, quotaRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}
		callListQuotas(requirementsFactory, quotaRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestListQuotas", func() {

		quota := models.QuotaFields{}
		quota.Name = "quota-name"
		quota.MemoryLimit = 1024

		quotaRepo := &testapi.FakeQuotaRepository{FindAllQuotas: []models.QuotaFields{quota}}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		ui := callListQuotas(requirementsFactory, quotaRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting quotas as", "my-user"},
			{"OK"},
			{"name", "memory limit"},
			{"quota-name", "1g"},
		})
	})
})
