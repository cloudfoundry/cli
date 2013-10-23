package user_test

import (
	"cf"
	. "cf/commands/user"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestOrgUsersFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	userRepo := &testapi.FakeUserRepository{}
	ui := callOrgUsers(t, []string{}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callOrgUsers(t, []string{"Org1"}, reqFactory, userRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestOrgUsersRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	userRepo := &testapi.FakeUserRepository{}
	args := []string{"Org1"}

	reqFactory.LoginSuccess = false
	callOrgUsers(t, args, reqFactory, userRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callOrgUsers(t, args, reqFactory, userRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, "Org1", reqFactory.OrganizationName)
}

func TestOrgUsers(t *testing.T) {
	org := cf.Organization{Name: "Found Org"}

	userRepo := &testapi.FakeUserRepository{}
	userRepo.FindAllInOrgByRoleUsersByRole = map[string][]cf.User{
		"MANAGER": []cf.User{cf.User{Username: "user1"}, cf.User{Username: "user2"}},
		"DEV":     []cf.User{cf.User{Username: "user3"}},
	}

	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess: true,
		Organization: org,
	}

	ui := callOrgUsers(t, []string{"Org1"}, reqFactory, userRepo)

	assert.Contains(t, ui.Outputs[0], "Getting users in org")
	assert.Contains(t, ui.Outputs[0], "Found Org")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, org, userRepo.FindAllInOrgByRoleOrganization)

	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[3], "MANAGER")
	assert.Contains(t, ui.Outputs[4], "user1")
	assert.Contains(t, ui.Outputs[5], "user2")

	assert.Contains(t, ui.Outputs[7], "DEV")
	assert.Contains(t, ui.Outputs[8], "user3")
}

func callOrgUsers(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewOrgUsers(ui, config, userRepo)
	ctxt := testcmd.NewContext("org-users", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
