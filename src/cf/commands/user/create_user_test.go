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

func getCreateUserDefaults() (defaultArgs []string, defaultReqs *testreq.FakeReqFactory, defaultUserRepo *testapi.FakeUserRepository) {
	defaultArgs = []string{"my-user", "my-password"}
	defaultReqs = &testreq.FakeReqFactory{LoginSuccess: true}
	defaultUserRepo = &testapi.FakeUserRepository{}
	return
}

func TestCreateUserFailsWithUsage(t *testing.T) {
	defaultArgs, defaultReqs, defaultUserRepo := getCreateUserDefaults()

	emptyArgs := []string{}

	ui := callCreateUser(t, emptyArgs, defaultReqs, defaultUserRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateUser(t, defaultArgs, defaultReqs, defaultUserRepo)
	assert.False(t, ui.FailedWithUsage)
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

	ui := callCreateUser(t, defaultArgs, defaultReqs, defaultUserRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating user", "my-user", "current-user"},
		{"OK"},
		{"TIP"},
	})
	assert.Equal(t, defaultUserRepo.CreateUserUsername, "my-user")
}

func TestCreateUserWhenItAlreadyExists(t *testing.T) {
	defaultArgs, defaultReqs, userAlreadyExistsRepo := getCreateUserDefaults()

	userAlreadyExistsRepo.CreateUserExists = true

	ui := callCreateUser(t, defaultArgs, defaultReqs, userAlreadyExistsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating user"},
		{"FAILED"},
		{"my-user"},
		{"already exists"},
	})
}

func callCreateUser(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-user", args)

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

	cmd := NewCreateUser(ui, config, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
