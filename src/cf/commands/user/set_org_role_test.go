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

func TestSetOrgRoleFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	userRepo := &testapi.FakeUserRepository{}

	ui := callSetOrgRole(t, []string{"my-user", "my-org", "my-role"}, reqFactory, userRepo)
	assert.False(t, ui.FailedWithUsage)

	ui = callSetOrgRole(t, []string{"my-user", "my-org"}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetOrgRole(t, []string{"my-user"}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetOrgRole(t, []string{}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)
}

func TestSetOrgRoleRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	userRepo := &testapi.FakeUserRepository{}

	reqFactory.LoginSuccess = false
	callSetOrgRole(t, []string{"my-user", "my-org", "my-role"}, reqFactory, userRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callSetOrgRole(t, []string{"my-user", "my-org", "my-role"}, reqFactory, userRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.UserUsername, "my-user")
	assert.Equal(t, reqFactory.OrganizationName, "my-org")
}

func TestSetOrgRole(t *testing.T) {
	org := cf.Organization{}
	org.Guid = "my-org-guid"
	org.Name = "my-org"
	user := cf.UserFields{}
	user.Guid = "my-user-guid"
	user.Username = "my-user"
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess: true,
		UserFields:   user,
		Organization: org,
	}
	userRepo := &testapi.FakeUserRepository{}

	ui := callSetOrgRole(t, []string{"some-user", "some-org", "OrgManager"}, reqFactory, userRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Assigning role", "OrgManager", "my-user", "my-org", "current-user"},
		{"OK"},
	})
	assert.Equal(t, userRepo.SetOrgRoleUserGuid, "my-user-guid")
	assert.Equal(t, userRepo.SetOrgRoleOrganizationGuid, "my-org-guid")
	assert.Equal(t, userRepo.SetOrgRoleRole, cf.ORG_MANAGER)
}

func callSetOrgRole(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-org-role", args)

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

	cmd := NewSetOrgRole(ui, config, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
