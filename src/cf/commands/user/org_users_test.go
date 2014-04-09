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

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

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

func callOrgUsers(args []string, requirementsFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewOrgUsers(ui, config, userRepo)
	ctxt := testcmd.NewContext("org-users", args)

	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}

var _ = Describe("Listing users in an org", func() {
	It("TestOrgUsersFailsWithUsage", func() {
		requirementsFactory := &testreq.FakeReqFactory{}
		userRepo := &testapi.FakeUserRepository{}
		ui := callOrgUsers([]string{}, requirementsFactory, userRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callOrgUsers([]string{"Org1"}, requirementsFactory, userRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})

	It("TestOrgUsersRequirements", func() {
		requirementsFactory := &testreq.FakeReqFactory{}
		userRepo := &testapi.FakeUserRepository{}
		args := []string{"Org1"}

		requirementsFactory.LoginSuccess = false
		callOrgUsers(args, requirementsFactory, userRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory.LoginSuccess = true
		callOrgUsers(args, requirementsFactory, userRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		Expect("Org1").To(Equal(requirementsFactory.OrganizationName))
	})

	It("TestOrgUsers", func() {
		org := models.Organization{}
		org.Name = "Found Org"
		org.Guid = "found-org-guid"

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
			models.ORG_MANAGER:     []models.UserFields{user, user2},
			models.BILLING_MANAGER: []models.UserFields{user4},
			models.ORG_AUDITOR:     []models.UserFields{user3},
		}

		requirementsFactory := &testreq.FakeReqFactory{
			LoginSuccess: true,
			Organization: org,
		}

		ui := callOrgUsers([]string{"Org1"}, requirementsFactory, userRepo)

		Expect(userRepo.ListUsersOrganizationGuid).To(Equal("found-org-guid"))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting users in org", "Found Org", "my-user"},
			{"ORG MANAGER"},
			{"user1"},
			{"user2"},
			{"BILLING MANAGER"},
			{"user4"},
			{"ORG AUDITOR"},
			{"user3"},
		})
	})

	It("lists all org users", func() {
		org := models.Organization{}
		org.Name = "Found Org"
		org.Guid = "found-org-guid"

		userRepo := &testapi.FakeUserRepository{}
		user := models.UserFields{}
		user.Username = "user1"
		user2 := models.UserFields{}
		user2.Username = "user2"
		userRepo.ListUsersByRole = map[string][]models.UserFields{
			models.ORG_USER: []models.UserFields{user, user2},
		}

		requirementsFactory := &testreq.FakeReqFactory{
			LoginSuccess: true,
			Organization: org,
		}

		ui := callOrgUsers([]string{"-a", "Org1"}, requirementsFactory, userRepo)

		Expect(userRepo.ListUsersOrganizationGuid).To(Equal("found-org-guid"))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting users in org", "Found Org", "my-user"},
			{"USERS"},
			{"user1"},
			{"user2"},
		})
	})
})
