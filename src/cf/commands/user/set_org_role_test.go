package user_test

import (
	. "cf/commands/user"
	"cf/configuration"
	"cf/models"
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

func callSetOrgRole(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-org-role", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "current-user",
	})
	assert.NoError(t, err)
	org2 := models.OrganizationFields{}
	org2.Name = "my-org"
	space := models.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org2,
		AccessToken:        token,
	}

	cmd := NewSetOrgRole(ui, config, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSetOrgRoleFailsWithUsage", func() {
			reqFactory := &testreq.FakeReqFactory{}
			userRepo := &testapi.FakeUserRepository{}

			ui := callSetOrgRole(mr.T(), []string{"my-user", "my-org", "my-role"}, reqFactory, userRepo)
			assert.False(mr.T(), ui.FailedWithUsage)

			ui = callSetOrgRole(mr.T(), []string{"my-user", "my-org"}, reqFactory, userRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callSetOrgRole(mr.T(), []string{"my-user"}, reqFactory, userRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callSetOrgRole(mr.T(), []string{}, reqFactory, userRepo)
			assert.True(mr.T(), ui.FailedWithUsage)
		})
		It("TestSetOrgRoleRequirements", func() {

			reqFactory := &testreq.FakeReqFactory{}
			userRepo := &testapi.FakeUserRepository{}

			reqFactory.LoginSuccess = false
			callSetOrgRole(mr.T(), []string{"my-user", "my-org", "my-role"}, reqFactory, userRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory.LoginSuccess = true
			callSetOrgRole(mr.T(), []string{"my-user", "my-org", "my-role"}, reqFactory, userRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			assert.Equal(mr.T(), reqFactory.UserUsername, "my-user")
			assert.Equal(mr.T(), reqFactory.OrganizationName, "my-org")
		})
		It("TestSetOrgRole", func() {

			org := models.Organization{}
			org.Guid = "my-org-guid"
			org.Name = "my-org"
			user := models.UserFields{}
			user.Guid = "my-user-guid"
			user.Username = "my-user"
			reqFactory := &testreq.FakeReqFactory{
				LoginSuccess: true,
				UserFields:   user,
				Organization: org,
			}
			userRepo := &testapi.FakeUserRepository{}

			ui := callSetOrgRole(mr.T(), []string{"some-user", "some-org", "OrgManager"}, reqFactory, userRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Assigning role", "OrgManager", "my-user", "my-org", "current-user"},
				{"OK"},
			})
			assert.Equal(mr.T(), userRepo.SetOrgRoleUserGuid, "my-user-guid")
			assert.Equal(mr.T(), userRepo.SetOrgRoleOrganizationGuid, "my-org-guid")
			assert.Equal(mr.T(), userRepo.SetOrgRoleRole, models.ORG_MANAGER)
		})
	})
}
