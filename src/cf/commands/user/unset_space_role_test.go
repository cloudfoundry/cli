package user_test

import (
	. "cf/commands/user"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewUnsetSpaceRole(ui, config, spaceRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestUnsetSpaceRoleFailsWithUsage", func() {
		reqFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()

		ui := callUnsetSpaceRole(mr.T(), []string{}, spaceRepo, userRepo, reqFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnsetSpaceRole(mr.T(), []string{"username"}, spaceRepo, userRepo, reqFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnsetSpaceRole(mr.T(), []string{"username", "org"}, spaceRepo, userRepo, reqFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnsetSpaceRole(mr.T(), []string{"username", "org", "space"}, spaceRepo, userRepo, reqFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnsetSpaceRole(mr.T(), []string{"username", "org", "space", "role"}, spaceRepo, userRepo, reqFactory)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestUnsetSpaceRoleRequirements", func() {

		reqFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()
		args := []string{"username", "org", "space", "role"}

		reqFactory.LoginSuccess = false
		callUnsetSpaceRole(mr.T(), args, spaceRepo, userRepo, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory.LoginSuccess = true
		callUnsetSpaceRole(mr.T(), args, spaceRepo, userRepo, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		Expect(reqFactory.UserUsername).To(Equal("username"))
		Expect(reqFactory.OrganizationName).To(Equal("org"))
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

		Expect(spaceRepo.FindByNameInOrgName).To(Equal("my-space"))
		Expect(spaceRepo.FindByNameInOrgOrgGuid).To(Equal("some-org-guid"))

		println(ui.DumpOutputs())
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Removing role", "SpaceManager", "some-user", "some-org", "some-space", "my-user"},
			{"OK"},
		})
		Expect(userRepo.UnsetSpaceRoleRole).To(Equal(models.SPACE_MANAGER))
		Expect(userRepo.UnsetSpaceRoleUserGuid).To(Equal("some-user-guid"))
		Expect(userRepo.UnsetSpaceRoleSpaceGuid).To(Equal("some-space-guid"))
	})
})
