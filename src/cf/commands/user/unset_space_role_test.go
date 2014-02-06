package user_test

import (
	. "cf/commands/user"
	"cf/configuration"
	"cf/models"
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

func getUnsetSpaceRoleDeps() (reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) {
	reqFactory = &testreq.FakeReqFactory{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	userRepo = &testapi.FakeUserRepository{}
	return
}

func callUnsetSpaceRole(t mr.TestingT, args []string, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("unset-space-role", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "current-user",
	})
	assert.NoError(t, err)
	space2 := models.SpaceFields{}
	space2.Name = "my-space"
	org2 := models.OrganizationFields{}
	org2.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space2,
		OrganizationFields: org2,
		AccessToken:        token,
	}

	cmd := NewUnsetSpaceRole(ui, config, spaceRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestUnsetSpaceRoleFailsWithUsage", func() {
			reqFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()

			ui := callUnsetSpaceRole(mr.T(), []string{}, spaceRepo, userRepo, reqFactory)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUnsetSpaceRole(mr.T(), []string{"username"}, spaceRepo, userRepo, reqFactory)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUnsetSpaceRole(mr.T(), []string{"username", "org"}, spaceRepo, userRepo, reqFactory)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUnsetSpaceRole(mr.T(), []string{"username", "org", "space"}, spaceRepo, userRepo, reqFactory)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUnsetSpaceRole(mr.T(), []string{"username", "org", "space", "role"}, spaceRepo, userRepo, reqFactory)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestUnsetSpaceRoleRequirements", func() {

			reqFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()
			args := []string{"username", "org", "space", "role"}

			reqFactory.LoginSuccess = false
			callUnsetSpaceRole(mr.T(), args, spaceRepo, userRepo, reqFactory)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory.LoginSuccess = true
			callUnsetSpaceRole(mr.T(), args, spaceRepo, userRepo, reqFactory)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			assert.Equal(mr.T(), reqFactory.UserUsername, "username")
			assert.Equal(mr.T(), reqFactory.OrganizationName, "org")
		})
		It("TestUnsetSpaceRole", func() {

			user := models.UserFields{}
			user.Username = "some-user"
			user.Guid = "some-user-guid"
			org := models.Organization{}
			org.Name = "some-org"
			org.Guid = "some-org-guid"

			reqFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()
			reqFactory.LoginSuccess = true
			reqFactory.UserFields = user
			reqFactory.Organization = org
			spaceRepo.FindByNameInOrgSpace = models.Space{}
			spaceRepo.FindByNameInOrgSpace.Name = "some-space"
			spaceRepo.FindByNameInOrgSpace.Guid = "some-space-guid"

			args := []string{"my-username", "my-org", "my-space", "SpaceManager"}

			ui := callUnsetSpaceRole(mr.T(), args, spaceRepo, userRepo, reqFactory)

			assert.Equal(mr.T(), spaceRepo.FindByNameInOrgName, "my-space")
			assert.Equal(mr.T(), spaceRepo.FindByNameInOrgOrgGuid, "some-org-guid")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Removing role", "SpaceManager", "some-user", "some-org", "some-space", "current-user"},
				{"OK"},
			})
			assert.Equal(mr.T(), userRepo.UnsetSpaceRoleRole, models.SPACE_MANAGER)
			assert.Equal(mr.T(), userRepo.UnsetSpaceRoleUserGuid, "some-user-guid")
			assert.Equal(mr.T(), userRepo.UnsetSpaceRoleSpaceGuid, "some-space-guid")
		})
	})
}
