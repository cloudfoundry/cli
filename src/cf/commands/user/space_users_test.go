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

func callSpaceUsers(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
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

	cmd := NewSpaceUsers(ui, config, spaceRepo, userRepo)
	ctxt := testcmd.NewContext("space-users", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSpaceUsersFailsWithUsage", func() {
			reqFactory := &testreq.FakeReqFactory{}
			spaceRepo := &testapi.FakeSpaceRepository{}
			userRepo := &testapi.FakeUserRepository{}

			ui := callSpaceUsers(mr.T(), []string{}, reqFactory, spaceRepo, userRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callSpaceUsers(mr.T(), []string{"my-org"}, reqFactory, spaceRepo, userRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callSpaceUsers(mr.T(), []string{"my-org", "my-space"}, reqFactory, spaceRepo, userRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestSpaceUsersRequirements", func() {

			reqFactory := &testreq.FakeReqFactory{}
			spaceRepo := &testapi.FakeSpaceRepository{}
			userRepo := &testapi.FakeUserRepository{}
			args := []string{"my-org", "my-space"}

			reqFactory.LoginSuccess = false
			callSpaceUsers(mr.T(), args, reqFactory, spaceRepo, userRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory.LoginSuccess = true
			callSpaceUsers(mr.T(), args, reqFactory, spaceRepo, userRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			assert.Equal(mr.T(), "my-org", reqFactory.OrganizationName)
		})
		It("TestSpaceUsers", func() {

			org := cf.Organization{}
			org.Name = "Org1"
			org.Guid = "org1-guid"
			space := cf.Space{}
			space.Name = "Space1"
			space.Guid = "space1-guid"

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}
			spaceRepo := &testapi.FakeSpaceRepository{FindByNameInOrgSpace: space}
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
				cf.SPACE_MANAGER:   []cf.UserFields{user, user2},
				cf.SPACE_DEVELOPER: []cf.UserFields{user4},
				cf.SPACE_AUDITOR:   []cf.UserFields{user3},
			}

			ui := callSpaceUsers(mr.T(), []string{"my-org", "my-space"}, reqFactory, spaceRepo, userRepo)

			assert.Equal(mr.T(), spaceRepo.FindByNameInOrgName, "my-space")
			assert.Equal(mr.T(), spaceRepo.FindByNameInOrgOrgGuid, "org1-guid")
			assert.Equal(mr.T(), userRepo.ListUsersSpaceGuid, "space1-guid")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting users in org", "Org1", "Space1", "my-user"},
				{"SPACE MANAGER"},
				{"user1"},
				{"user2"},
				{"SPACE DEVELOPER"},
				{"user4"},
				{"SPACE AUDITOR"},
				{"user3"},
			})
		})
	})
}
