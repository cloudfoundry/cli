package organization_test

import (
	. "cf/commands/organization"
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

func callCreateOrg(args []string, reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (fakeUI *testterm.FakeUI) {
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

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestCreateOrgFailsWithUsage", func() {
		orgRepo := &testapi.FakeOrgRepository{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		ui := callCreateOrg([]string{}, reqFactory, orgRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateOrg([]string{"my-org"}, reqFactory, orgRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestCreateOrgRequirements", func() {

		orgRepo := &testapi.FakeOrgRepository{}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		callCreateOrg([]string{"my-org"}, reqFactory, orgRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
		callCreateOrg([]string{"my-org"}, reqFactory, orgRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestCreateOrg", func() {

		orgRepo := &testapi.FakeOrgRepository{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		ui := callCreateOrg([]string{"my-org"}, reqFactory, orgRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating org", "my-org", "my-user"},
			{"OK"},
		})
		Expect(orgRepo.CreateName).To(Equal("my-org"))
	})
	It("TestCreateOrgWhenAlreadyExists", func() {

		orgRepo := &testapi.FakeOrgRepository{CreateOrgExists: true}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		ui := callCreateOrg([]string{"my-org"}, reqFactory, orgRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating org", "my-org"},
			{"OK"},
			{"my-org", "already exists"},
		})
	})
})
