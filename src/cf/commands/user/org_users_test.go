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
	org := cf.Organization{}
	org.Name = "Found Org"
	org.Guid = "found-org-guid"

	userRepo := &testapi.FakeUserRepository{}
	user := cf.UserFields{}
	user.Username = "user1"
	user2 := cf.UserFields{}
	user2.Username = "user2"
	user3 := cf.UserFields{}
	user3.Username = "user3"
	user4 := cf.UserFields{}
	user4.Username = "user4"
	userRepo.ListUsersByRole = map[string][]cf.UserFields{
		cf.ORG_MANAGER:     []cf.UserFields{user, user2},
		cf.BILLING_MANAGER: []cf.UserFields{user4},
		cf.ORG_AUDITOR:     []cf.UserFields{user3},
	}

	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess: true,
		Organization: org,
	}

	ui := callOrgUsers(t, []string{"Org1"}, reqFactory, userRepo)

	assert.Equal(t, userRepo.ListUsersOrganizationGuid, "found-org-guid")

	assert.Contains(t, ui.Outputs[0], "Getting users in org")
	assert.Contains(t, ui.Outputs[0], "Found Org")
	assert.Contains(t, ui.Outputs[0], "my-user")

	testassert.SliceContains(t, ui.Outputs, []string{
		"ORG MANAGER",
		"user1",
		"user2",
		"BILLING MANAGER",
		"user4",
		"ORG AUDITOR",
		"user3",
	})
}

func callOrgUsers(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org3 := cf.OrganizationFields{}
	org3.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org3,
		AccessToken:        token,
	}

	cmd := NewOrgUsers(ui, config, userRepo)
	ctxt := testcmd.NewContext("org-users", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
