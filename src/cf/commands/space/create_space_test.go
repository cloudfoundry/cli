package space_test

import (
	. "cf/commands/space"
	"cf/commands/user"
	"cf/models"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	"testhelpers/maker"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var (
	defaultReqFactory *testreq.FakeReqFactory
	configSpace       models.SpaceFields
	configOrg         models.OrganizationFields
	defaultSpace      models.Space
	defaultSpaceRepo  *testapi.FakeSpaceRepository
	defaultOrgRepo    *testapi.FakeOrgRepository
	defaultUserRepo   *testapi.FakeUserRepository
)

func resetSpaceDefaults() {
	defaultReqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	configOrg = models.OrganizationFields{}
	configOrg.Name = "my-org"
	configOrg.Guid = "my-org-guid"

	configSpace = models.SpaceFields{}
	configSpace.Name = "config-space"
	configSpace.Guid = "config-space-guid"

	defaultSpace = models.Space{}
	defaultSpace.Name = "my-space"
	defaultSpace.Guid = "my-space-guid"
	defaultSpace.Organization = configOrg

	defaultSpaceRepo = &testapi.FakeSpaceRepository{
		CreateSpaceSpace: defaultSpace,
	}

	defaultUserRepo = &testapi.FakeUserRepository{}
	defaultOrgRepo = &testapi.FakeOrgRepository{}
}

func callCreateSpace(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, orgRepo *testapi.FakeOrgRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-space", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	spaceRoleSetter := user.NewSetSpaceRole(ui, configRepo, spaceRepo, userRepo)
	cmd := NewCreateSpace(ui, configRepo, spaceRoleSetter, spaceRepo, orgRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestCreateSpaceFailsWithUsage", func() {
		resetSpaceDefaults()
		reqFactory := &testreq.FakeReqFactory{}

		ui := callCreateSpace(mr.T(), []string{}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callCreateSpace(mr.T(), []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
		assert.False(mr.T(), ui.FailedWithUsage)
	})
	It("TestCreateSpaceRequirements", func() {

		resetSpaceDefaults()
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		callCreateSpace(mr.T(), []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callCreateSpace(mr.T(), []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
		callCreateSpace(mr.T(), []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callCreateSpace(mr.T(), []string{"-o", "some-org", "my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)
	})
	It("TestCreateSpace", func() {

		resetSpaceDefaults()
		ui := callCreateSpace(mr.T(), []string{"my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating space", "my-space", "my-org", "my-user"},
			{"OK"},
			{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_MANAGER]},
			{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_DEVELOPER]},
			{"TIP"},
		})

		assert.Equal(mr.T(), defaultSpaceRepo.CreateSpaceName, "my-space")
		assert.Equal(mr.T(), defaultSpaceRepo.CreateSpaceOrgGuid, "my-org-guid")
		assert.Equal(mr.T(), defaultUserRepo.SetSpaceRoleUserGuid, "my-user-guid")
		assert.Equal(mr.T(), defaultUserRepo.SetSpaceRoleSpaceGuid, "my-space-guid")
		assert.Equal(mr.T(), defaultUserRepo.SetSpaceRoleRole, models.SPACE_DEVELOPER)
	})
	It("TestCreateSpaceInOrg", func() {

		resetSpaceDefaults()

		org := maker.NewOrg(maker.Overrides{"name": "other-org"})
		defaultOrgRepo.Organizations = []models.Organization{org}

		ui := callCreateSpace(mr.T(), []string{"-o", "other-org", "my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating space", "my-space", "other-org", "my-user"},
			{"OK"},
			{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_MANAGER]},
			{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_DEVELOPER]},
			{"TIP"},
		})

		assert.Equal(mr.T(), defaultSpaceRepo.CreateSpaceName, "my-space")
		assert.Equal(mr.T(), defaultSpaceRepo.CreateSpaceOrgGuid, org.Guid)
		assert.Equal(mr.T(), defaultUserRepo.SetSpaceRoleUserGuid, "my-user-guid")
		assert.Equal(mr.T(), defaultUserRepo.SetSpaceRoleSpaceGuid, "my-space-guid")
		assert.Equal(mr.T(), defaultUserRepo.SetSpaceRoleRole, models.SPACE_DEVELOPER)
	})
	It("TestCreateSpaceInOrgWhenTheOrgDoesNotExist", func() {

		resetSpaceDefaults()

		defaultOrgRepo.FindByNameNotFound = true

		ui := callCreateSpace(mr.T(), []string{"-o", "cool-organization", "my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"cool-organization", "does not exist"},
		})

		assert.Equal(mr.T(), defaultSpaceRepo.CreateSpaceName, "")
	})
	It("TestCreateSpaceInOrgWhenErrorFindingOrg", func() {

		resetSpaceDefaults()

		defaultOrgRepo.FindByNameErr = true

		ui := callCreateSpace(mr.T(), []string{"-o", "cool-organization", "my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"error"},
		})

		assert.Equal(mr.T(), defaultSpaceRepo.CreateSpaceName, "")
	})
	It("TestCreateSpaceWhenItAlreadyExists", func() {

		resetSpaceDefaults()
		defaultSpaceRepo.CreateSpaceExists = true
		ui := callCreateSpace(mr.T(), []string{"my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating space", "my-space"},
			{"OK"},
			{"my-space", "already exists"},
		})
		testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
			{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_MANAGER]},
		})

		assert.Equal(mr.T(), defaultSpaceRepo.CreateSpaceName, "")
		assert.Equal(mr.T(), defaultSpaceRepo.CreateSpaceOrgGuid, "")
		assert.Equal(mr.T(), defaultUserRepo.SetSpaceRoleUserGuid, "")
		assert.Equal(mr.T(), defaultUserRepo.SetSpaceRoleSpaceGuid, "")
	})
})
