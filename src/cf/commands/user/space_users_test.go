package user_test

import (
	. "cf/commands/user"
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

func callSpaceUsers(args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewSpaceUsers(ui, config, spaceRepo, userRepo)
	ctxt := testcmd.NewContext("space-users", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestSpaceUsersFailsWithUsage", func() {
		reqFactory := &testreq.FakeReqFactory{}
		spaceRepo := &testapi.FakeSpaceRepository{}
		userRepo := &testapi.FakeUserRepository{}

		ui := callSpaceUsers([]string{}, reqFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSpaceUsers([]string{"my-org"}, reqFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSpaceUsers([]string{"my-org", "my-space"}, reqFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestSpaceUsersRequirements", func() {

		reqFactory := &testreq.FakeReqFactory{}
		spaceRepo := &testapi.FakeSpaceRepository{}
		userRepo := &testapi.FakeUserRepository{}
		args := []string{"my-org", "my-space"}

		reqFactory.LoginSuccess = false
		callSpaceUsers(args, reqFactory, spaceRepo, userRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory.LoginSuccess = true
		callSpaceUsers(args, reqFactory, spaceRepo, userRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		Expect("my-org").To(Equal(reqFactory.OrganizationName))
	})
	It("TestSpaceUsers", func() {

		org := models.Organization{}
		org.Name = "Org1"
		org.Guid = "org1-guid"
		space := models.Space{}
		space.Name = "Space1"
		space.Guid = "space1-guid"

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}
		spaceRepo := &testapi.FakeSpaceRepository{FindByNameInOrgSpace: space}
		userRepo := &testapi.FakeUserRepository{}

		user := models.UserFields{}
		user.Username = "user1"
		user2 := models.UserFields{}
		user2.Username = "user2"
		user3 := models.UserFields{}
		user3.Username = "user3"
		user4 := models.UserFields{}
		user4.Username = "user4"
		userRepo.ListUsersByRole = map[string][]models.UserFields{
			models.SPACE_MANAGER:   []models.UserFields{user, user2},
			models.SPACE_DEVELOPER: []models.UserFields{user4},
			models.SPACE_AUDITOR:   []models.UserFields{user3},
		}

		ui := callSpaceUsers([]string{"my-org", "my-space"}, reqFactory, spaceRepo, userRepo)

		Expect(spaceRepo.FindByNameInOrgName).To(Equal("my-space"))
		Expect(spaceRepo.FindByNameInOrgOrgGuid).To(Equal("org1-guid"))
		Expect(userRepo.ListUsersSpaceGuid).To(Equal("space1-guid"))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
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
