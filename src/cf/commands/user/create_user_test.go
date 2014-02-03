package user_test

import (
	"cf"
	. "cf/commands/user"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func getCreateUserDefaults() (defaultArgs []string, defaultReqs *testreq.FakeReqFactory, defaultUserRepo *testapi.FakeUserRepository) {
	defaultArgs = []string{"my-user", "my-password"}
	defaultReqs = &testreq.FakeReqFactory{LoginSuccess: true}
	defaultUserRepo = &testapi.FakeUserRepository{}
	return
}

func callCreateUser(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestCreateUserFailsWithUsage", func() {
			defaultArgs, defaultReqs, defaultUserRepo := getCreateUserDefaults()

			emptyArgs := []string{}

			ui := callCreateUser(mr.T(), emptyArgs, defaultReqs, defaultUserRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callCreateUser(mr.T(), defaultArgs, defaultReqs, defaultUserRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestCreateUserRequirements", func() {

			defaultArgs, defaultReqs, defaultUserRepo := getCreateUserDefaults()

			callCreateUser(mr.T(), defaultArgs, defaultReqs, defaultUserRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			notLoggedInReq := &testreq.FakeReqFactory{LoginSuccess: false}
			callCreateUser(mr.T(), defaultArgs, notLoggedInReq, defaultUserRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestCreateUser", func() {

			defaultArgs, defaultReqs, defaultUserRepo := getCreateUserDefaults()

			ui := callCreateUser(mr.T(), defaultArgs, defaultReqs, defaultUserRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating user", "my-user", "current-user"},
				{"OK"},
				{"TIP"},
			})
			assert.Equal(mr.T(), defaultUserRepo.CreateUserUsername, "my-user")
		})
		It("TestCreateUserWhenItAlreadyExists", func() {

			defaultArgs, defaultReqs, userAlreadyExistsRepo := getCreateUserDefaults()

			userAlreadyExistsRepo.CreateUserExists = true

			ui := callCreateUser(mr.T(), defaultArgs, defaultReqs, userAlreadyExistsRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating user"},
				{"FAILED"},
				{"my-user"},
				{"already exists"},
			})
		})
	})
}
