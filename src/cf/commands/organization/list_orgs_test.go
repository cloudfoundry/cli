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

func callListOrgs(config configuration.Reader, requirementsFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("orgs", []string{})
	cmd := organization.NewListOrgs(fakeUI, config, orgRepo)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {

	It("TestListOrgsRequirements", func() {
		orgRepo := &testapi.FakeOrgRepository{}
		config := testconfig.NewRepository()

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		callListOrgs(config, requirementsFactory, orgRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}
		callListOrgs(config, requirementsFactory, orgRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("TestListAllPagesOfOrgs", func() {
		org1 := models.Organization{}
		org1.Name = "Organization-1"

		org2 := models.Organization{}
		org2.Name = "Organization-2"

		org3 := models.Organization{}
		org3.Name = "Organization-3"

		orgRepo := &testapi.FakeOrgRepository{
			Organizations: []models.Organization{org1, org2, org3},
		}

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		tokenInfo := configuration.TokenInfo{Username: "my-user"}
		config := testconfig.NewRepositoryWithAccessToken(tokenInfo)

		ui := callListOrgs(config, requirementsFactory, orgRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting orgs as my-user"},
			{"Organization-1"},
			{"Organization-2"},
			{"Organization-3"},
		})
	})

	It("TestListNoOrgs", func() {
		orgs := []models.Organization{}
		orgRepo := &testapi.FakeOrgRepository{
			Organizations: orgs,
		}

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		tokenInfo := configuration.TokenInfo{Username: "my-user"}
		config := testconfig.NewRepositoryWithAccessToken(tokenInfo)

		ui := callListOrgs(config, requirementsFactory, orgRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting orgs as my-user"},
			{"No orgs found"},
		})
	})
})
