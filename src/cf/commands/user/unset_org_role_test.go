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

func callUnsetOrgRole(t mr.TestingT, args []string, userRepo *testapi.FakeUserRepository, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("unset-org-role", args)

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

	cmd := NewUnsetOrgRole(ui, config, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestUnsetOrgRoleFailsWithUsage", func() {
			userRepo := &testapi.FakeUserRepository{}
			reqFactory := &testreq.FakeReqFactory{}

			ui := callUnsetOrgRole(mr.T(), []string{}, userRepo, reqFactory)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUnsetOrgRole(mr.T(), []string{"username"}, userRepo, reqFactory)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUnsetOrgRole(mr.T(), []string{"username", "org"}, userRepo, reqFactory)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUnsetOrgRole(mr.T(), []string{"username", "org", "role"}, userRepo, reqFactory)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestUnsetOrgRoleRequirements", func() {

			userRepo := &testapi.FakeUserRepository{}
			reqFactory := &testreq.FakeReqFactory{}
			args := []string{"username", "org", "role"}

			reqFactory.LoginSuccess = false
			callUnsetOrgRole(mr.T(), args, userRepo, reqFactory)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory.LoginSuccess = true
			callUnsetOrgRole(mr.T(), args, userRepo, reqFactory)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			assert.Equal(mr.T(), reqFactory.UserUsername, "username")
			assert.Equal(mr.T(), reqFactory.OrganizationName, "org")
		})
		It("TestUnsetOrgRole", func() {

			userRepo := &testapi.FakeUserRepository{}
			user := models.UserFields{}
			user.Username = "some-user"
			user.Guid = "some-user-guid"
			org := models.Organization{}
			org.Name = "some-org"
			org.Guid = "some-org-guid"
			reqFactory := &testreq.FakeReqFactory{
				LoginSuccess: true,
				UserFields:   user,
				Organization: org,
			}
			args := []string{"my-username", "my-org", "OrgManager"}

			ui := callUnsetOrgRole(mr.T(), args, userRepo, reqFactory)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Removing role", "OrgManager", "my-username", "my-org", "current-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), userRepo.UnsetOrgRoleRole, models.ORG_MANAGER)
			assert.Equal(mr.T(), userRepo.UnsetOrgRoleUserGuid, "some-user-guid")
			assert.Equal(mr.T(), userRepo.UnsetOrgRoleOrganizationGuid, "some-org-guid")
		})
	})
}
