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

func callSetOrgRole(args []string, requirementsFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-org-role", args)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewSetOrgRole(ui, config, userRepo)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestSetOrgRoleFailsWithUsage", func() {
		requirementsFactory := &testreq.FakeReqFactory{}
		userRepo := &testapi.FakeUserRepository{}

		ui := callSetOrgRole([]string{"my-user", "my-org", "my-role"}, requirementsFactory, userRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())

		ui = callSetOrgRole([]string{"my-user", "my-org"}, requirementsFactory, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSetOrgRole([]string{"my-user"}, requirementsFactory, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSetOrgRole([]string{}, requirementsFactory, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("TestSetOrgRoleRequirements", func() {
		requirementsFactory := &testreq.FakeReqFactory{}
		userRepo := &testapi.FakeUserRepository{}

		requirementsFactory.LoginSuccess = false
		callSetOrgRole([]string{"my-user", "my-org", "my-role"}, requirementsFactory, userRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory.LoginSuccess = true
		callSetOrgRole([]string{"my-user", "my-org", "my-role"}, requirementsFactory, userRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		Expect(requirementsFactory.UserUsername).To(Equal("my-user"))
		Expect(requirementsFactory.OrganizationName).To(Equal("my-org"))
	})

	It("TestSetOrgRole", func() {
		org := models.Organization{}
		org.Guid = "my-org-guid"
		org.Name = "my-org"
		user := models.UserFields{}
		user.Guid = "my-user-guid"
		user.Username = "my-user"
		requirementsFactory := &testreq.FakeReqFactory{
			LoginSuccess: true,
			UserFields:   user,
			Organization: org,
		}
		userRepo := &testapi.FakeUserRepository{}

		ui := callSetOrgRole([]string{"some-user", "some-org", "OrgManager"}, requirementsFactory, userRepo)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Assigning role", "OrgManager", "my-user", "my-org", "my-user"},
			[]string{"OK"},
		))
		Expect(userRepo.SetOrgRoleUserGuid).To(Equal("my-user-guid"))
		Expect(userRepo.SetOrgRoleOrganizationGuid).To(Equal("my-org-guid"))
		Expect(userRepo.SetOrgRoleRole).To(Equal(models.ORG_MANAGER))
	})
})
