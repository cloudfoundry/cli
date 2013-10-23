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

func getCreateUserDefaults() (defaultArgs []string, defaultReqs *testreq.FakeReqFactory, defaultUserRepo *testapi.FakeUserRepository) {
	defaultArgs = []string{"my-user", "my-password"}
	defaultReqs = &testreq.FakeReqFactory{LoginSuccess: true}
	defaultUserRepo = &testapi.FakeUserRepository{}
	return
}

func TestCreateUserFailsWithUsage(t *testing.T) {
	defaultArgs, defaultReqs, defaultUserRepo := getCreateUserDefaults()

	emptyArgs := []string{}

	fakeUI := callCreateUser(t, emptyArgs, defaultReqs, defaultUserRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callCreateUser(t, defaultArgs, defaultReqs, defaultUserRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func TestCreateUserRequirements(t *testing.T) {
	defaultArgs, defaultReqs, defaultUserRepo := getCreateUserDefaults()

	callCreateUser(t, defaultArgs, defaultReqs, defaultUserRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	notLoggedInReq := &testreq.FakeReqFactory{LoginSuccess: false}
	callCreateUser(t, defaultArgs, notLoggedInReq, defaultUserRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

}

func TestCreateUser(t *testing.T) {
	defaultArgs, defaultReqs, defaultUserRepo := getCreateUserDefaults()

	fakeUI := callCreateUser(t, defaultArgs, defaultReqs, defaultUserRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating user")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")
	assert.Contains(t, fakeUI.Outputs[0], "current-user")
	assert.Equal(t, defaultUserRepo.CreateUserUser.Username, "my-user")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "TIP")
}

func TestCreateUserWhenItAlreadyExists(t *testing.T) {
	defaultArgs, defaultReqs, userAlreadyExistsRepo := getCreateUserDefaults()

	userAlreadyExistsRepo.CreateUserExists = true

	fakeUI := callCreateUser(t, defaultArgs, defaultReqs, userAlreadyExistsRepo)

	assert.Equal(t, len(fakeUI.Outputs), 3)
	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
	assert.Contains(t, fakeUI.Outputs[2], "my-user")
}

func callCreateUser(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-user", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "current-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewCreateUser(ui, config, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
