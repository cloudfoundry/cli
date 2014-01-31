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

func callDeleteUser(t mr.TestingT, args []string, userRepo *testapi.FakeUserRepository, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
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

func deleteWithConfirmation(t mr.TestingT, confirmation string) (ui *testterm.FakeUI, userRepo *testapi.FakeUserRepository) {
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestDeleteUserFailsWithUsage", func() {
			userRepo := &testapi.FakeUserRepository{}
			reqFactory := &testreq.FakeReqFactory{}

			ui := callDeleteUser(mr.T(), []string{}, userRepo, reqFactory)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callDeleteUser(mr.T(), []string{"foo"}, userRepo, reqFactory)
			assert.False(mr.T(), ui.FailedWithUsage)

			ui = callDeleteUser(mr.T(), []string{"foo", "bar"}, userRepo, reqFactory)
			assert.True(mr.T(), ui.FailedWithUsage)
		})
		It("TestDeleteUserRequirements", func() {

			userRepo := &testapi.FakeUserRepository{}
			reqFactory := &testreq.FakeReqFactory{}
			args := []string{"-f", "my-user"}

			reqFactory.LoginSuccess = false
			callDeleteUser(mr.T(), args, userRepo, reqFactory)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory.LoginSuccess = true
			callDeleteUser(mr.T(), args, userRepo, reqFactory)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestDeleteUserWhenConfirmingWithY", func() {

			ui, userRepo := deleteWithConfirmation(mr.T(), "Y")

			assert.Equal(mr.T(), len(ui.Outputs), 2)
			assert.Equal(mr.T(), len(ui.Prompts), 1)
			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"Really delete", "my-user"},
			})
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting user", "my-user", "current-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), userRepo.FindByUsernameUsername, "my-user")
			assert.Equal(mr.T(), userRepo.DeleteUserGuid, "my-found-user-guid")
		})
		It("TestDeleteUserWhenConfirmingWithYes", func() {

			ui, userRepo := deleteWithConfirmation(mr.T(), "Yes")

			assert.Equal(mr.T(), len(ui.Outputs), 2)
			assert.Equal(mr.T(), len(ui.Prompts), 1)
			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"Really delete", "my-user"},
			})
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting user", "my-user", "current-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), userRepo.FindByUsernameUsername, "my-user")
			assert.Equal(mr.T(), userRepo.DeleteUserGuid, "my-found-user-guid")
		})
		It("TestDeleteUserWhenNotConfirming", func() {

			ui, userRepo := deleteWithConfirmation(mr.T(), "Nope")

			assert.Equal(mr.T(), len(ui.Outputs), 0)
			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{{"Really delete"}})

			assert.Equal(mr.T(), userRepo.FindByUsernameUsername, "")
			assert.Equal(mr.T(), userRepo.DeleteUserGuid, "")
		})
		It("TestDeleteUserWithForceOption", func() {

			foundUserFields := cf.UserFields{}
			foundUserFields.Guid = "my-found-user-guid"
			userRepo := &testapi.FakeUserRepository{FindByUsernameUserFields: foundUserFields}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			ui := callDeleteUser(mr.T(), []string{"-f", "my-user"}, userRepo, reqFactory)

			assert.Equal(mr.T(), len(ui.Outputs), 2)
			assert.Equal(mr.T(), len(ui.Prompts), 0)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting user", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), userRepo.FindByUsernameUsername, "my-user")
			assert.Equal(mr.T(), userRepo.DeleteUserGuid, "my-found-user-guid")
		})
		It("TestDeleteUserWhenUserNotFound", func() {

			userRepo := &testapi.FakeUserRepository{FindByUsernameNotFound: true}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			ui := callDeleteUser(mr.T(), []string{"-f", "my-user"}, userRepo, reqFactory)

			assert.Equal(mr.T(), len(ui.Outputs), 3)
			assert.Equal(mr.T(), len(ui.Prompts), 0)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting user", "my-user"},
				{"OK"},
				{"my-user", "does not exist"},
			})

			assert.Equal(mr.T(), userRepo.FindByUsernameUsername, "my-user")
			assert.Equal(mr.T(), userRepo.DeleteUserGuid, "")
		})
	})
}
