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

func getSetSpaceRoleDeps() (reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) {
	reqFactory = &testreq.FakeReqFactory{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	userRepo = &testapi.FakeUserRepository{}
	return
}

func callSetSpaceRole(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-space-role", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "current-user",
	})
	assert.NoError(t, err)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"

	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewSetSpaceRole(ui, config, spaceRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSetSpaceRoleFailsWithUsage", func() {
			reqFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()

			ui := callSetSpaceRole(mr.T(), []string{}, reqFactory, spaceRepo, userRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callSetSpaceRole(mr.T(), []string{"my-user"}, reqFactory, spaceRepo, userRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callSetSpaceRole(mr.T(), []string{"my-user", "my-org"}, reqFactory, spaceRepo, userRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callSetSpaceRole(mr.T(), []string{"my-user", "my-org", "my-space"}, reqFactory, spaceRepo, userRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callSetSpaceRole(mr.T(), []string{"my-user", "my-org", "my-space", "my-role"}, reqFactory, spaceRepo, userRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestSetSpaceRoleRequirements", func() {

			args := []string{"username", "org", "space", "role"}
			reqFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()

			reqFactory.LoginSuccess = false
			callSetSpaceRole(mr.T(), args, reqFactory, spaceRepo, userRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory.LoginSuccess = true
			callSetSpaceRole(mr.T(), args, reqFactory, spaceRepo, userRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			assert.Equal(mr.T(), reqFactory.UserUsername, "username")
			assert.Equal(mr.T(), reqFactory.OrganizationName, "org")
		})
		It("TestSetSpaceRole", func() {

			org := cf.Organization{}
			org.Guid = "my-org-guid"
			org.Name = "my-org"

			args := []string{"some-user", "some-org", "some-space", "SpaceManager"}

			reqFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()
			reqFactory.LoginSuccess = true
			reqFactory.UserFields = cf.UserFields{}
			reqFactory.UserFields.Guid = "my-user-guid"
			reqFactory.UserFields.Username = "my-user"
			reqFactory.Organization = org

			spaceRepo.FindByNameInOrgSpace = cf.Space{}
			spaceRepo.FindByNameInOrgSpace.Guid = "my-space-guid"
			spaceRepo.FindByNameInOrgSpace.Name = "my-space"
			spaceRepo.FindByNameInOrgSpace.Organization = org.OrganizationFields

			ui := callSetSpaceRole(mr.T(), args, reqFactory, spaceRepo, userRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Assigning role ", "SpaceManager", "my-user", "my-org", "my-space", "current-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), spaceRepo.FindByNameInOrgName, "some-space")
			assert.Equal(mr.T(), spaceRepo.FindByNameInOrgOrgGuid, "my-org-guid")

			assert.Equal(mr.T(), userRepo.SetSpaceRoleUserGuid, "my-user-guid")
			assert.Equal(mr.T(), userRepo.SetSpaceRoleSpaceGuid, "my-space-guid")
			assert.Equal(mr.T(), userRepo.SetSpaceRoleRole, cf.SPACE_MANAGER)
		})
	})
}
