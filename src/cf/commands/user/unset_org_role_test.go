package user_test

import (
	"cf"
	. "cf/commands/user"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestUnsetOrgRoleFailsWithUsage(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}
	reqFactory := &testreq.FakeReqFactory{}

	ui := callUnsetOrgRole(t, []string{}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetOrgRole(t, []string{"username"}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetOrgRole(t, []string{"username", "org"}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetOrgRole(t, []string{"username", "org", "role"}, userRepo, reqFactory)
	assert.False(t, ui.FailedWithUsage)
}

func TestUnsetOrgRoleRequirements(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}
	reqFactory := &testreq.FakeReqFactory{}
	args := []string{"username", "org", "role"}

	reqFactory.LoginSuccess = false
	callUnsetOrgRole(t, args, userRepo, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callUnsetOrgRole(t, args, userRepo, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.UserUsername, "username")
	assert.Equal(t, reqFactory.OrganizationName, "org")
}

func TestUnsetOrgRole(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}
	user := cf.UserFields{}
	user.Username = "some-user"
	user.Guid = "some-user-guid"
	org := cf.Organization{}
	org.Name = "some-org"
	org.Guid = "some-org-guid"
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess: true,
		UserFields:   user,
		Organization: org,
	}
	args := []string{"my-username", "my-org", "OrgManager"}

	ui := callUnsetOrgRole(t, args, userRepo, reqFactory)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Removing role", "OrgManager", "my-username", "my-org", "current-user"},
		{"OK"},
	})

	assert.Equal(t, userRepo.UnsetOrgRoleRole, cf.ORG_MANAGER)
	assert.Equal(t, userRepo.UnsetOrgRoleUserGuid, "some-user-guid")
	assert.Equal(t, userRepo.UnsetOrgRoleOrganizationGuid, "some-org-guid")
}

func callUnsetOrgRole(t *testing.T, args []string, userRepo *testapi.FakeUserRepository, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("unset-org-role", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "current-user",
	})
	assert.NoError(t, err)
	org2 := cf.OrganizationFields{}
	org2.Name = "my-org"
	space := cf.SpaceFields{}
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
