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

func getUnsetSpaceRoleDeps() (requirementsFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) {
	requirementsFactory = &testreq.FakeReqFactory{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	userRepo = &testapi.FakeUserRepository{}
	return
}

func callUnsetSpaceRole(args []string, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository, requirementsFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	config := testconfig.NewRepositoryWithDefaults()
	cmd := NewUnsetSpaceRole(ui, config, spaceRepo, userRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestUnsetSpaceRoleFailsWithUsage", func() {
		requirementsFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()

		ui := callUnsetSpaceRole([]string{}, spaceRepo, userRepo, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnsetSpaceRole([]string{"username"}, spaceRepo, userRepo, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnsetSpaceRole([]string{"username", "org"}, spaceRepo, userRepo, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnsetSpaceRole([]string{"username", "org", "space"}, spaceRepo, userRepo, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnsetSpaceRole([]string{"username", "org", "space", "role"}, spaceRepo, userRepo, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestUnsetSpaceRoleRequirements", func() {

		requirementsFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()
		args := []string{"username", "org", "space", "role"}

		requirementsFactory.LoginSuccess = false
		callUnsetSpaceRole(args, spaceRepo, userRepo, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory.LoginSuccess = true
		callUnsetSpaceRole(args, spaceRepo, userRepo, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		Expect(requirementsFactory.UserUsername).To(Equal("username"))
		Expect(requirementsFactory.OrganizationName).To(Equal("org"))
	})

	It("TestUnsetSpaceRole", func() {
		user := models.UserFields{}
		user.Username = "some-user"
		user.Guid = "some-user-guid"
		org := models.Organization{}
		org.Name = "some-org"
		org.Guid = "some-org-guid"

		requirementsFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()
		requirementsFactory.LoginSuccess = true
		requirementsFactory.UserFields = user
		requirementsFactory.Organization = org
		spaceRepo.FindByNameInOrgSpace = models.Space{}
		spaceRepo.FindByNameInOrgSpace.Name = "some-space"
		spaceRepo.FindByNameInOrgSpace.Guid = "some-space-guid"

		args := []string{"my-username", "my-org", "my-space", "SpaceManager"}

		ui := callUnsetSpaceRole(args, spaceRepo, userRepo, requirementsFactory)

		Expect(spaceRepo.FindByNameInOrgName).To(Equal("my-space"))
		Expect(spaceRepo.FindByNameInOrgOrgGuid).To(Equal("some-org-guid"))

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Removing role", "SpaceManager", "some-user", "some-org", "some-space", "my-user"},
			[]string{"OK"},
		))
		Expect(userRepo.UnsetSpaceRoleRole).To(Equal(models.SPACE_MANAGER))
		Expect(userRepo.UnsetSpaceRoleUserGuid).To(Equal("some-user-guid"))
		Expect(userRepo.UnsetSpaceRoleSpaceGuid).To(Equal("some-space-guid"))
	})
})
