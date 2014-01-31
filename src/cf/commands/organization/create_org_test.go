package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callCreateOrg(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-org", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	space := cf.SpaceFields{}
	space.Name = "my-space"

	organization := cf.OrganizationFields{}
	organization.Name = "my-org"

	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: organization,
		AccessToken:        token,
	}

	cmd := NewCreateOrg(fakeUI, config, orgRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestCreateOrgFailsWithUsage", func() {
			orgRepo := &testapi.FakeOrgRepository{}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			ui := callCreateOrg(mr.T(), []string{}, reqFactory, orgRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callCreateOrg(mr.T(), []string{"my-org"}, reqFactory, orgRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestCreateOrgRequirements", func() {

			orgRepo := &testapi.FakeOrgRepository{}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			callCreateOrg(mr.T(), []string{"my-org"}, reqFactory, orgRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			callCreateOrg(mr.T(), []string{"my-org"}, reqFactory, orgRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestCreateOrg", func() {

			orgRepo := &testapi.FakeOrgRepository{}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			ui := callCreateOrg(mr.T(), []string{"my-org"}, reqFactory, orgRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating org", "my-org", "my-user"},
				{"OK"},
			})
			assert.Equal(mr.T(), orgRepo.CreateName, "my-org")
		})
		It("TestCreateOrgWhenAlreadyExists", func() {

			orgRepo := &testapi.FakeOrgRepository{CreateOrgExists: true}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			ui := callCreateOrg(mr.T(), []string{"my-org"}, reqFactory, orgRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating org", "my-org"},
				{"OK"},
				{"my-org", "already exists"},
			})
		})
	})
}
