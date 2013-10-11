package user_test

import (
	. "cf/commands/user"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func getCreateUserDefaults() (defaultArgs []string, defaultReqs *testreq.FakeReqFactory, defaultUserRepo *testapi.FakeUserRepository) {
	defaultArgs = []string{"my-user", "my-password"}
	defaultReqs = &testreq.FakeReqFactory{LoginSuccess: true}
	defaultUserRepo = &testapi.FakeUserRepository{}
	return
}

func TestCreateUserFailsWithUsage(t *testing.T) {
	defaultArgs, defaultReqs, defaultUserRepo := getCreateUserDefaults()

	emptyArgs := []string{}

	fakeUI := callCreateUser(emptyArgs, defaultReqs, defaultUserRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callCreateUser(defaultArgs, defaultReqs, defaultUserRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func TestCreateUserRequirements(t *testing.T) {
	defaultArgs, defaultReqs, defaultUserRepo := getCreateUserDefaults()

	callCreateUser(defaultArgs, defaultReqs, defaultUserRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	notLoggedInReq := &testreq.FakeReqFactory{LoginSuccess: false}
	callCreateUser(defaultArgs, notLoggedInReq, defaultUserRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

}

func TestCreateUser(t *testing.T) {
	defaultArgs, defaultReqs, defaultUserRepo := getCreateUserDefaults()

	fakeUI := callCreateUser(defaultArgs, defaultReqs, defaultUserRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating user")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")
	assert.Equal(t, defaultUserRepo.CreateUserUser.Username, "my-user")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "TIP")
}

func TestCreateUserWhenItAlreadyExists(t *testing.T) {
	defaultArgs, defaultReqs, userAlreadyExistsRepo := getCreateUserDefaults()

	userAlreadyExistsRepo.CreateUserExists = true

	fakeUI := callCreateUser(defaultArgs, defaultReqs, userAlreadyExistsRepo)

	assert.Equal(t, len(fakeUI.Outputs), 3)
	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
	assert.Contains(t, fakeUI.Outputs[2], "my-user")
}

func callCreateUser(args []string, reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-user", args)

	cmd := NewCreateUser(ui, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
