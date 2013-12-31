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

func TestDeleteUserFailsWithUsage(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}
	reqFactory := &testreq.FakeReqFactory{}

	ui := callDeleteUser(t, []string{}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callDeleteUser(t, []string{"foo"}, userRepo, reqFactory)
	assert.False(t, ui.FailedWithUsage)

	ui = callDeleteUser(t, []string{"foo", "bar"}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)
}

func TestDeleteUserRequirements(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}
	reqFactory := &testreq.FakeReqFactory{}
	args := []string{"-f", "my-user"}

	reqFactory.LoginSuccess = false
	callDeleteUser(t, args, userRepo, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callDeleteUser(t, args, userRepo, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestDeleteUserWhenConfirmingWithY(t *testing.T) {
	ui, userRepo := deleteWithConfirmation(t, "Y")

	assert.Equal(t, len(ui.Outputs), 2)
	assert.Equal(t, len(ui.Prompts), 1)
	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"Really delete", "my-user"},
	})
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting user", "my-user", "current-user"},
		{"OK"},
	})

	assert.Equal(t, userRepo.FindByUsernameUsername, "my-user")
	assert.Equal(t, userRepo.DeleteUserGuid, "my-found-user-guid")
}

func TestDeleteUserWhenConfirmingWithYes(t *testing.T) {
	ui, userRepo := deleteWithConfirmation(t, "Yes")

	assert.Equal(t, len(ui.Outputs), 2)
	assert.Equal(t, len(ui.Prompts), 1)
	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"Really delete", "my-user"},
	})
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting user", "my-user", "current-user"},
		{"OK"},
	})

	assert.Equal(t, userRepo.FindByUsernameUsername, "my-user")
	assert.Equal(t, userRepo.DeleteUserGuid, "my-found-user-guid")
}

func TestDeleteUserWhenNotConfirming(t *testing.T) {
	ui, userRepo := deleteWithConfirmation(t, "Nope")

	assert.Equal(t, len(ui.Outputs), 0)
	testassert.SliceContains(t, ui.Prompts, testassert.Lines{{"Really delete"}})

	assert.Equal(t, userRepo.FindByUsernameUsername, "")
	assert.Equal(t, userRepo.DeleteUserGuid, "")
}

func TestDeleteUserWithForceOption(t *testing.T) {
	foundUserFields := cf.UserFields{}
	foundUserFields.Guid = "my-found-user-guid"
	userRepo := &testapi.FakeUserRepository{FindByUsernameUserFields: foundUserFields}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	ui := callDeleteUser(t, []string{"-f", "my-user"}, userRepo, reqFactory)

	assert.Equal(t, len(ui.Outputs), 2)
	assert.Equal(t, len(ui.Prompts), 0)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting user", "my-user"},
		{"OK"},
	})

	assert.Equal(t, userRepo.FindByUsernameUsername, "my-user")
	assert.Equal(t, userRepo.DeleteUserGuid, "my-found-user-guid")
}

func TestDeleteUserWhenUserNotFound(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{FindByUsernameNotFound: true}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	ui := callDeleteUser(t, []string{"-f", "my-user"}, userRepo, reqFactory)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Equal(t, len(ui.Prompts), 0)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting user", "my-user"},
		{"OK"},
		{"my-user", "does not exist"},
	})

	assert.Equal(t, userRepo.FindByUsernameUsername, "my-user")
	assert.Equal(t, userRepo.DeleteUserGuid, "")
}

func callDeleteUser(t *testing.T, args []string, userRepo *testapi.FakeUserRepository, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "current-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewDeleteUser(ui, config, userRepo)
	ctxt := testcmd.NewContext("delete-user", args)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func deleteWithConfirmation(t *testing.T, confirmation string) (ui *testterm.FakeUI, userRepo *testapi.FakeUserRepository) {
	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}
	user2 := cf.UserFields{}
	user2.Username = "my-found-user"
	user2.Guid = "my-found-user-guid"
	userRepo = &testapi.FakeUserRepository{
		FindByUsernameUserFields: user2,
	}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "current-user",
	})
	assert.NoError(t, err)
	org2 := cf.OrganizationFields{}
	org2.Name = "my-org"
	space2 := cf.SpaceFields{}
	space2.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space2,
		OrganizationFields: org2,
		AccessToken:        token,
	}

	cmd := NewDeleteUser(ui, config, userRepo)

	ctxt := testcmd.NewContext("delete-user", []string{"my-user"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
