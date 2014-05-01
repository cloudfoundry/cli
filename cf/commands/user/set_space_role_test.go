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
	"github.com/cloudfoundry/cli/cf/configuration"
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

func getSetSpaceRoleDeps() (requirementsFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) {
	requirementsFactory = &testreq.FakeReqFactory{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	userRepo = &testapi.FakeUserRepository{}
	return
}

func callSetSpaceRole(args []string, requirementsFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-space-role", args)
	configRepo := testconfig.NewRepositoryWithDefaults()
	accessToken, err := testconfig.EncodeAccessToken(configuration.TokenInfo{
		Username: "current-user",
	})
	Expect(err).NotTo(HaveOccurred())
	configRepo.SetAccessToken(accessToken)

	cmd := NewSetSpaceRole(ui, configRepo, spaceRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestSetSpaceRoleFailsWithUsage", func() {
		requirementsFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()

		ui := callSetSpaceRole([]string{}, requirementsFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSetSpaceRole([]string{"my-user"}, requirementsFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSetSpaceRole([]string{"my-user", "my-org"}, requirementsFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSetSpaceRole([]string{"my-user", "my-org", "my-space"}, requirementsFactory, spaceRepo, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSetSpaceRole([]string{"my-user", "my-org", "my-space", "my-role"}, requirementsFactory, spaceRepo, userRepo)

		Expect(ui.FailedWithUsage).To(BeFalse())
	})

	It("TestSetSpaceRoleRequirements", func() {
		args := []string{"username", "org", "space", "role"}
		requirementsFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()

		requirementsFactory.LoginSuccess = false
		callSetSpaceRole(args, requirementsFactory, spaceRepo, userRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory.LoginSuccess = true
		callSetSpaceRole(args, requirementsFactory, spaceRepo, userRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		Expect(requirementsFactory.UserUsername).To(Equal("username"))
		Expect(requirementsFactory.OrganizationName).To(Equal("org"))
	})

	It("TestSetSpaceRole", func() {
		org := models.Organization{}
		org.Guid = "my-org-guid"
		org.Name = "my-org"

		args := []string{"some-user", "some-org", "some-space", "SpaceManager"}

		requirementsFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()
		requirementsFactory.LoginSuccess = true
		requirementsFactory.UserFields = models.UserFields{}
		requirementsFactory.UserFields.Guid = "my-user-guid"
		requirementsFactory.UserFields.Username = "my-user"
		requirementsFactory.Organization = org

		spaceRepo.FindByNameInOrgSpace = models.Space{}
		spaceRepo.FindByNameInOrgSpace.Guid = "my-space-guid"
		spaceRepo.FindByNameInOrgSpace.Name = "my-space"
		spaceRepo.FindByNameInOrgSpace.Organization = org.OrganizationFields

		ui := callSetSpaceRole(args, requirementsFactory, spaceRepo, userRepo)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Assigning role ", "SpaceManager", "my-user", "my-org", "my-space", "current-user"},
			[]string{"OK"},
		))

		Expect(spaceRepo.FindByNameInOrgName).To(Equal("some-space"))
		Expect(spaceRepo.FindByNameInOrgOrgGuid).To(Equal("my-org-guid"))

		Expect(userRepo.SetSpaceRoleUserGuid).To(Equal("my-user-guid"))
		Expect(userRepo.SetSpaceRoleSpaceGuid).To(Equal("my-space-guid"))
		Expect(userRepo.SetSpaceRoleRole).To(Equal(models.SPACE_MANAGER))
	})
})
