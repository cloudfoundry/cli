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

var _ = Describe("Testing with ginkgo", func() {
	It("TestUnsetOrgRoleFailsWithUsage", func() {
		userRepo := &testapi.FakeUserRepository{}
		requirementsFactory := &testreq.FakeReqFactory{}

		ui := callUnsetOrgRole([]string{}, userRepo, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnsetOrgRole([]string{"username"}, userRepo, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnsetOrgRole([]string{"username", "org"}, userRepo, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnsetOrgRole([]string{"username", "org", "role"}, userRepo, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestUnsetOrgRoleRequirements", func() {

		userRepo := &testapi.FakeUserRepository{}
		requirementsFactory := &testreq.FakeReqFactory{}
		args := []string{"username", "org", "role"}

		requirementsFactory.LoginSuccess = false
		callUnsetOrgRole(args, userRepo, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory.LoginSuccess = true
		callUnsetOrgRole(args, userRepo, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		Expect(requirementsFactory.UserUsername).To(Equal("username"))
		Expect(requirementsFactory.OrganizationName).To(Equal("org"))
	})
	It("TestUnsetOrgRole", func() {

		userRepo := &testapi.FakeUserRepository{}
		user := models.UserFields{}
		user.Username = "some-user"
		user.Guid = "some-user-guid"
		org := models.Organization{}
		org.Name = "some-org"
		org.Guid = "some-org-guid"
		requirementsFactory := &testreq.FakeReqFactory{
			LoginSuccess: true,
			UserFields:   user,
			Organization: org,
		}
		args := []string{"my-username", "my-org", "OrgManager"}

		ui := callUnsetOrgRole(args, userRepo, requirementsFactory)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Removing role", "OrgManager", "my-username", "my-org", "my-user"},
			[]string{"OK"},
		))

		Expect(userRepo.UnsetOrgRoleRole).To(Equal(models.ORG_MANAGER))
		Expect(userRepo.UnsetOrgRoleUserGuid).To(Equal("some-user-guid"))
		Expect(userRepo.UnsetOrgRoleOrganizationGuid).To(Equal("some-org-guid"))
	})
})

func callUnsetOrgRole(args []string, userRepo *testapi.FakeUserRepository, requirementsFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewUnsetOrgRole(ui, configRepo, userRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}
