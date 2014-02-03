package organization_test

import (
	"cf"
	"cf/commands/organization"
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

func callRenameOrg(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("rename-org", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	spaceFields := cf.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := cf.OrganizationFields{}
	orgFields.Name = "my-org"

	config := &configuration.Configuration{
		SpaceFields:        spaceFields,
		OrganizationFields: orgFields,
		AccessToken:        token,
	}

	cmd := organization.NewRenameOrg(ui, config, orgRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestRenameOrgFailsWithUsage", func() {
			reqFactory := &testreq.FakeReqFactory{}
			orgRepo := &testapi.FakeOrgRepository{}

			ui := callRenameOrg(mr.T(), []string{}, reqFactory, orgRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callRenameOrg(mr.T(), []string{"foo"}, reqFactory, orgRepo)
			assert.True(mr.T(), ui.FailedWithUsage)
		})
		It("TestRenameOrgRequirements", func() {

			orgRepo := &testapi.FakeOrgRepository{}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			callRenameOrg(mr.T(), []string{"my-org", "my-new-org"}, reqFactory, orgRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			assert.Equal(mr.T(), reqFactory.OrganizationName, "my-org")
		})
		It("TestRenameOrgRun", func() {

			orgRepo := &testapi.FakeOrgRepository{}

			org := cf.Organization{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}
			ui := callRenameOrg(mr.T(), []string{"my-org", "my-new-org"}, reqFactory, orgRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Renaming org", "my-org", "my-new-org", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), orgRepo.RenameOrganizationGuid, "my-org-guid")
			assert.Equal(mr.T(), orgRepo.RenameNewName, "my-new-org")
		})
	})
}
