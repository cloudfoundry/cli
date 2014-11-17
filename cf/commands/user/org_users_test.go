package user_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("org-users command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.ReadWriter
		userRepo            *testapi.FakeUserRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		userRepo = &testapi.FakeUserRepository{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewOrgUsers(ui, configRepo, userRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when invoked without an org name", func() {
			requirementsFactory.LoginSuccess = true

			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			Expect(runCommand("say-hello-to-my-little-org")).To(BeFalse())
		})
	})

	Context("when logged in and given an org with users", func() {
		BeforeEach(func() {
			org := models.Organization{}
			org.Name = "the-org"
			org.Guid = "the-org-guid"

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

			requirementsFactory.LoginSuccess = true
			requirementsFactory.Organization = org
		})

		It("shows the special users in the given org", func() {
			runCommand("the-org")

			Expect(userRepo.ListUsersOrganizationGuid).To(Equal("the-org-guid"))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting users in org", "the-org", "my-user"},
				[]string{"ORG MANAGER"},
				[]string{"user1"},
				[]string{"user2"},
				[]string{"BILLING MANAGER"},
				[]string{"user4"},
				[]string{"ORG AUDITOR"},
				[]string{"user3"},
			))
		})

		Context("when the -a flag is provided", func() {
			BeforeEach(func() {
				user := models.UserFields{}
				user.Username = "user1"
				user2 := models.UserFields{}
				user2.Username = "user2"
				userRepo.ListUsersByRole = map[string][]models.UserFields{
					models.ORG_USER: []models.UserFields{user, user2},
				}
			})

			It("lists all org users, regardless of role", func() {
				runCommand("-a", "the-org")

				Expect(userRepo.ListUsersOrganizationGuid).To(Equal("the-org-guid"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting users in org", "the-org", "my-user"},
					[]string{"USERS"},
					[]string{"user1"},
					[]string{"user2"},
				))
			})
		})
	})
})
