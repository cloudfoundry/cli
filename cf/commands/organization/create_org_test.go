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
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

func callCreateOrg(args []string, requirementsFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-org", args)

	space := models.SpaceFields{}
	space.Name = "my-space"

	organization := models.OrganizationFields{}
	organization.Name = "my-org"

	token := configuration.TokenInfo{Username: "my-user"}
	config := testconfig.NewRepositoryWithAccessToken(token)
	config.SetSpaceFields(space)
	config.SetOrganizationFields(organization)

	cmd := NewCreateOrg(fakeUI, config, orgRepo)

	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestCreateOrgFailsWithUsage", func() {
		orgRepo := &testapi.FakeOrgRepository{}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		ui := callCreateOrg([]string{}, requirementsFactory, orgRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateOrg([]string{"my-org"}, requirementsFactory, orgRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})

	It("TestCreateOrgRequirements", func() {
		orgRepo := &testapi.FakeOrgRepository{}

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		callCreateOrg([]string{"my-org"}, requirementsFactory, orgRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}
		callCreateOrg([]string{"my-org"}, requirementsFactory, orgRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("TestCreateOrg", func() {
		orgRepo := &testapi.FakeOrgRepository{}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		ui := callCreateOrg([]string{"my-org"}, requirementsFactory, orgRepo)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating org", "my-org", "my-user"},
			[]string{"OK"},
		))
		Expect(orgRepo.CreateName).To(Equal("my-org"))
	})

	It("TestCreateOrgWhenAlreadyExists", func() {
		orgRepo := &testapi.FakeOrgRepository{CreateOrgExists: true}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		ui := callCreateOrg([]string{"my-org"}, requirementsFactory, orgRepo)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating org", "my-org"},
			[]string{"OK"},
			[]string{"my-org", "already exists"},
		))
	})
})
