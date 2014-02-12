package space_test

import (
	. "cf/commands/space"
	"cf/commands/user"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

func callCreateSpace(args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, orgRepo *testapi.FakeOrgRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
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

		ui := callCreateSpace([]string{}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateSpace([]string{"my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestCreateSpaceRequirements", func() {

		resetSpaceDefaults()
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		callCreateSpace([]string{"my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callCreateSpace([]string{"my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
		callCreateSpace([]string{"my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callCreateSpace([]string{"-o", "some-org", "my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
	})
	It("TestCreateSpace", func() {

		resetSpaceDefaults()
		ui := callCreateSpace([]string{"my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating space", "my-space", "my-org", "my-user"},
			{"OK"},
			{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_MANAGER]},
			{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_DEVELOPER]},
			{"TIP"},
		})

		Expect(defaultSpaceRepo.CreateSpaceName).To(Equal("my-space"))
		Expect(defaultSpaceRepo.CreateSpaceOrgGuid).To(Equal("my-org-guid"))
		Expect(defaultUserRepo.SetSpaceRoleUserGuid).To(Equal("my-user-guid"))
		Expect(defaultUserRepo.SetSpaceRoleSpaceGuid).To(Equal("my-space-guid"))
		Expect(defaultUserRepo.SetSpaceRoleRole).To(Equal(models.SPACE_DEVELOPER))
	})
	It("TestCreateSpaceInOrg", func() {

		resetSpaceDefaults()

		org := maker.NewOrg(maker.Overrides{"name": "other-org"})
		defaultOrgRepo.Organizations = []models.Organization{org}

		ui := callCreateSpace([]string{"-o", "other-org", "my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating space", "my-space", "other-org", "my-user"},
			{"OK"},
			{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_MANAGER]},
			{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_DEVELOPER]},
			{"TIP"},
		})

		Expect(defaultSpaceRepo.CreateSpaceName).To(Equal("my-space"))
		Expect(defaultSpaceRepo.CreateSpaceOrgGuid).To(Equal(org.Guid))
		Expect(defaultUserRepo.SetSpaceRoleUserGuid).To(Equal("my-user-guid"))
		Expect(defaultUserRepo.SetSpaceRoleSpaceGuid).To(Equal("my-space-guid"))
		Expect(defaultUserRepo.SetSpaceRoleRole).To(Equal(models.SPACE_DEVELOPER))
	})
	It("TestCreateSpaceInOrgWhenTheOrgDoesNotExist", func() {

		resetSpaceDefaults()

		defaultOrgRepo.FindByNameNotFound = true

		ui := callCreateSpace([]string{"-o", "cool-organization", "my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"cool-organization", "does not exist"},
		})

		Expect(defaultSpaceRepo.CreateSpaceName).To(Equal(""))
	})
	It("TestCreateSpaceInOrgWhenErrorFindingOrg", func() {

		resetSpaceDefaults()

		defaultOrgRepo.FindByNameErr = true

		ui := callCreateSpace([]string{"-o", "cool-organization", "my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"error"},
		})

		Expect(defaultSpaceRepo.CreateSpaceName).To(Equal(""))
	})
	It("TestCreateSpaceWhenItAlreadyExists", func() {

		resetSpaceDefaults()
		defaultSpaceRepo.CreateSpaceExists = true
		ui := callCreateSpace([]string{"my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating space", "my-space"},
			{"OK"},
			{"my-space", "already exists"},
		})
		testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
			{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_MANAGER]},
		})

		Expect(defaultSpaceRepo.CreateSpaceName).To(Equal(""))
		Expect(defaultSpaceRepo.CreateSpaceOrgGuid).To(Equal(""))
		Expect(defaultUserRepo.SetSpaceRoleUserGuid).To(Equal(""))
		Expect(defaultUserRepo.SetSpaceRoleSpaceGuid).To(Equal(""))
	})
})
