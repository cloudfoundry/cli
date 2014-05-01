/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package user_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

func callSpaceUsers(args []string, requirementsFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewSpaceUsers(ui, config, spaceRepo, userRepo)
	ctxt := testcmd.NewContext("space-users", args)

	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestSpaceUsersFailsWithUsage", func() {
		requirementsFactory := &testreq.FakeReqFactory{}
		spaceRepo := &testapi.FakeSpaceRepository{}
		userRepo := &testapi.FakeUserRepository{}

		ui := callSpaceUsers([]string{}, requirementsFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSpaceUsers([]string{"my-org"}, requirementsFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSpaceUsers([]string{"my-org", "my-space"}, requirementsFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestSpaceUsersRequirements", func() {

		requirementsFactory := &testreq.FakeReqFactory{}
		spaceRepo := &testapi.FakeSpaceRepository{}
		userRepo := &testapi.FakeUserRepository{}
		args := []string{"my-org", "my-space"}

		requirementsFactory.LoginSuccess = false
		callSpaceUsers(args, requirementsFactory, spaceRepo, userRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory.LoginSuccess = true
		callSpaceUsers(args, requirementsFactory, spaceRepo, userRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		Expect("my-org").To(Equal(requirementsFactory.OrganizationName))
	})
	It("TestSpaceUsers", func() {

		org := models.Organization{}
		org.Name = "Org1"
		org.Guid = "org1-guid"
		space := models.Space{}
		space.Name = "Space1"
		space.Guid = "space1-guid"

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}
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

		ui := callSpaceUsers([]string{"my-org", "my-space"}, requirementsFactory, spaceRepo, userRepo)

		Expect(spaceRepo.FindByNameInOrgName).To(Equal("my-space"))
		Expect(spaceRepo.FindByNameInOrgOrgGuid).To(Equal("org1-guid"))
		Expect(userRepo.ListUsersSpaceGuid).To(Equal("space1-guid"))

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting users in org", "Org1", "Space1", "my-user"},
			[]string{"SPACE MANAGER"},
			[]string{"user1"},
			[]string{"user2"},
			[]string{"SPACE DEVELOPER"},
			[]string{"user4"},
			[]string{"SPACE AUDITOR"},
			[]string{"user3"},
		))
	})
})
