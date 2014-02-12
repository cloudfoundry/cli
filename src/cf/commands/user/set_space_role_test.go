package user_test

import (
	. "cf/commands/user"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

func callSetSpaceRole(args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-space-role", args)
	configRepo := testconfig.NewRepositoryWithDefaults()
	accessToken, err := testconfig.EncodeAccessToken(configuration.TokenInfo{
		Username: "current-user",
	})
	Expect(err).NotTo(HaveOccurred())
	configRepo.SetAccessToken(accessToken)

	cmd := NewSetSpaceRole(ui, configRepo, spaceRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestSetSpaceRoleFailsWithUsage", func() {
		reqFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()

		ui := callSetSpaceRole([]string{}, reqFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSetSpaceRole([]string{"my-user"}, reqFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSetSpaceRole([]string{"my-user", "my-org"}, reqFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSetSpaceRole([]string{"my-user", "my-org", "my-space"}, reqFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSetSpaceRole([]string{"my-user", "my-org", "my-space", "my-role"}, reqFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestSetSpaceRoleRequirements", func() {

		args := []string{"username", "org", "space", "role"}
		reqFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()

		reqFactory.LoginSuccess = false
		callSetSpaceRole(args, reqFactory, spaceRepo, userRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory.LoginSuccess = true
		callSetSpaceRole(args, reqFactory, spaceRepo, userRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		Expect(reqFactory.UserUsername).To(Equal("username"))
		Expect(reqFactory.OrganizationName).To(Equal("org"))
	})
	It("TestSetSpaceRole", func() {

		org := models.Organization{}
		org.Guid = "my-org-guid"
		org.Name = "my-org"

		args := []string{"some-user", "some-org", "some-space", "SpaceManager"}

		reqFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()
		reqFactory.LoginSuccess = true
		reqFactory.UserFields = models.UserFields{}
		reqFactory.UserFields.Guid = "my-user-guid"
		reqFactory.UserFields.Username = "my-user"
		reqFactory.Organization = org

		spaceRepo.FindByNameInOrgSpace = models.Space{}
		spaceRepo.FindByNameInOrgSpace.Guid = "my-space-guid"
		spaceRepo.FindByNameInOrgSpace.Name = "my-space"
		spaceRepo.FindByNameInOrgSpace.Organization = org.OrganizationFields

		ui := callSetSpaceRole(args, reqFactory, spaceRepo, userRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Assigning role ", "SpaceManager", "my-user", "my-org", "my-space", "current-user"},
			{"OK"},
		})

		Expect(spaceRepo.FindByNameInOrgName).To(Equal("some-space"))
		Expect(spaceRepo.FindByNameInOrgOrgGuid).To(Equal("my-org-guid"))

		Expect(userRepo.SetSpaceRoleUserGuid).To(Equal("my-user-guid"))
		Expect(userRepo.SetSpaceRoleSpaceGuid).To(Equal("my-space-guid"))
		Expect(userRepo.SetSpaceRoleRole).To(Equal(models.SPACE_MANAGER))
	})
})
