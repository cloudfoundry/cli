package user_test

import (
	"cf"
	. "cf/commands/user"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestDeleteUserFailsWithUsage(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}
	reqFactory := &testreq.FakeReqFactory{}

	ui := callDeleteUser([]string{}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callDeleteUser([]string{"foo"}, userRepo, reqFactory)
	assert.False(t, ui.FailedWithUsage)

	ui = callDeleteUser([]string{"foo", "bar"}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)
}

func TestDeleteUserRequirements(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}
	reqFactory := &testreq.FakeReqFactory{}
	args := []string{"-f", "my-user"}

	reqFactory.LoginSuccess = false
	callDeleteUser(args, userRepo, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callDeleteUser(args, userRepo, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestDeleteUserWhenConfirmingWithY(t *testing.T) {
	ui, userRepo := deleteWithConfirmation("Y")

	assert.Equal(t, len(ui.Outputs), 2)
	assert.Equal(t, len(ui.Prompts), 1)
	assert.Contains(t, ui.Prompts[0], "Really delete")
	assert.Contains(t, ui.Outputs[0], "Deleting user")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, userRepo.FindByUsernameUsername, "my-user")
	assert.Equal(t, userRepo.DeleteUser.Guid, "my-found-user-guid")

	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteUserWhenConfirmingWithYes(t *testing.T) {
	ui, userRepo := deleteWithConfirmation("Yes")

	assert.Equal(t, len(ui.Outputs), 2)
	assert.Equal(t, len(ui.Prompts), 1)
	assert.Contains(t, ui.Prompts[0], "Really delete")
	assert.Contains(t, ui.Outputs[0], "Deleting user")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, userRepo.FindByUsernameUsername, "my-user")
	assert.Equal(t, userRepo.DeleteUser.Guid, "my-found-user-guid")

	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteUserWhenNotConfirming(t *testing.T) {
	ui, userRepo := deleteWithConfirmation("Nope")

	assert.Equal(t, len(ui.Outputs), 0)
	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Equal(t, userRepo.FindByUsernameUsername, "")
	assert.Equal(t, userRepo.DeleteUser.Guid, "")
}

func TestDeleteUserWithForceOption(t *testing.T) {
	foundUser := cf.User{Guid: "my-found-user-guid"}
	userRepo := &testapi.FakeUserRepository{FindByUsernameUser: foundUser}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	ui := callDeleteUser([]string{"-f", "my-user"}, userRepo, reqFactory)

	assert.Equal(t, len(ui.Outputs), 2)
	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting user")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, userRepo.FindByUsernameUsername, "my-user")
	assert.Equal(t, userRepo.DeleteUser.Guid, "my-found-user-guid")

	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteUserWhenUserNotFound(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{FindByUsernameNotFound: true}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	ui := callDeleteUser([]string{"-f", "my-user"}, userRepo, reqFactory)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting user")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, userRepo.FindByUsernameUsername, "my-user")
	assert.Equal(t, userRepo.DeleteUser.Guid, "")

	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "User not found")
}

func callDeleteUser(args []string, userRepo *testapi.FakeUserRepository, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	cmd := NewDeleteUser(ui, userRepo)
	ctxt := testcmd.NewContext("delete-user", args)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func deleteWithConfirmation(confirmation string) (ui *testterm.FakeUI, userRepo *testapi.FakeUserRepository) {
	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}

	userRepo = &testapi.FakeUserRepository{
		FindByUsernameUser: cf.User{Username: "my-found-user", Guid: "my-found-user-guid"},
	}

	cmd := NewDeleteUser(ui, userRepo)

	ctxt := testcmd.NewContext("delete-user", []string{"my-user"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
