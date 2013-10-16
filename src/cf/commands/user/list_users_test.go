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

func TestListUsersRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	userRepo := &testapi.FakeUserRepository{}

	reqFactory.LoginSuccess = false
	callListUsers(reqFactory, userRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callListUsers(reqFactory, userRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestListUsers(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	users := []cf.User{
		{Username: "user-1", IsAdmin: false},
		{Username: "user-2", IsAdmin: true},
	}
	userRepo := &testapi.FakeUserRepository{FindAllUsers: users}

	ui := callListUsers(reqFactory, userRepo)

	assert.Contains(t, ui.Outputs[0], "Getting users in all orgs and spaces")
	assert.True(t, userRepo.FindAllWasCalled)
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[3], "user-1")
	assert.Contains(t, ui.Outputs[4], "user-2")
	assert.Contains(t, ui.Outputs[4], "(admin)")
}

func callListUsers(reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	cmd := NewListUsers(ui, userRepo)
	ctxt := testcmd.NewContext("users", []string{})
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
