package user_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/plugin/models"
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
		configRepo          core_config.Repository
		userRepo            *testapi.FakeUserRepository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetUserRepository(userRepo)

		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("org-users").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		userRepo = &testapi.FakeUserRepository{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		deps = command_registry.NewDependency()
		updateCommandDependency(false)
	})

	runCommand := func(args ...string) bool {
		cmd := command_registry.Commands.FindCommand("org-users")
		return testcmd.RunCliCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when invoked without an org name", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
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
			updateCommandDependency(false)
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

		Context("when cc api verson is >= 2.21.0", func() {
			BeforeEach(func() {
				userRepo.ListUsersInOrgForRole_CallCount = 0
				userRepo.ListUsersInOrgForRoleWithNoUAA_CallCount = 0
			})

			It("calls ListUsersInOrgForRoleWithNoUAA()", func() {
				configRepo.SetApiVersion("2.22.0")
				runCommand("the-org")

				Expect(userRepo.ListUsersInOrgForRoleWithNoUAA_CallCount).To(BeNumerically(">=", 1))
				Expect(userRepo.ListUsersInOrgForRole_CallCount).To(Equal(0))
			})
		})

		Context("when cc api verson is < 2.21.0", func() {
			It("calls ListUsersInOrgForRole()", func() {
				configRepo.SetApiVersion("2.20.0")
				runCommand("the-org")

				Expect(userRepo.ListUsersInOrgForRoleWithNoUAA_CallCount).To(Equal(0))
				Expect(userRepo.ListUsersInOrgForRole_CallCount).To(BeNumerically(">=", 1))
			})
		})
	})

	Describe("when invoked by a plugin", func() {
		var (
			pluginUserModel []plugin_models.GetOrgUsers_Model
		)

		Context("single roles", func() {

			BeforeEach(func() {
				org := models.Organization{}
				org.Name = "the-org"
				org.Guid = "the-org-guid"

				// org managers
				user := models.UserFields{}
				user.Username = "user1"
				user.Guid = "1111"

				user2 := models.UserFields{}
				user2.Username = "user2"
				user2.Guid = "2222"

				// billing manager
				user3 := models.UserFields{}
				user3.Username = "user3"
				user3.Guid = "3333"

				// auditors
				user4 := models.UserFields{}
				user4.Username = "user4"
				user4.Guid = "4444"

				userRepo.ListUsersByRole = map[string][]models.UserFields{
					models.ORG_MANAGER:     []models.UserFields{user, user2},
					models.BILLING_MANAGER: []models.UserFields{user4},
					models.ORG_AUDITOR:     []models.UserFields{user3},
					models.ORG_USER:        []models.UserFields{user3},
				}

				requirementsFactory.LoginSuccess = true
				requirementsFactory.Organization = org
				pluginUserModel = []plugin_models.GetOrgUsers_Model{}
				deps.PluginModels.OrgUsers = &pluginUserModel
				updateCommandDependency(true)
			})

			It("populates the plugin model with users with single roles", func() {
				runCommand("the-org")
				Ω(pluginUserModel).To(HaveLen(4))

				for _, u := range pluginUserModel {
					switch u.Username {
					case "user1":
						Ω(u.Guid).To(Equal("1111"))
						Ω(u.Roles).To(ConsistOf([]string{models.ORG_MANAGER}))
					case "user2":
						Ω(u.Guid).To(Equal("2222"))
						Ω(u.Roles).To(ConsistOf([]string{models.ORG_MANAGER}))
					case "user3":
						Ω(u.Guid).To(Equal("3333"))
						Ω(u.Roles).To(ConsistOf([]string{models.ORG_AUDITOR}))
					case "user4":
						Ω(u.Guid).To(Equal("4444"))
						Ω(u.Roles).To(ConsistOf([]string{models.BILLING_MANAGER}))
					default:
						Fail("unexpected user: " + u.Username)
					}
				}

			})

			It("populates the plugin model with users with single roles -a flag", func() {
				runCommand("-a", "the-org")
				Ω(pluginUserModel).To(HaveLen(1))
				Ω(pluginUserModel[0].Username).To(Equal("user3"))
				Ω(pluginUserModel[0].Guid).To(Equal("3333"))
				Ω(pluginUserModel[0].Roles[0]).To(Equal(models.ORG_USER))
			})

		})

		Context("multiple roles", func() {

			BeforeEach(func() {
				org := models.Organization{}
				org.Name = "the-org"
				org.Guid = "the-org-guid"

				// org managers
				user := models.UserFields{}
				user.Username = "user1"
				user.Guid = "1111"
				user.IsAdmin = true

				user2 := models.UserFields{}
				user2.Username = "user2"
				user2.Guid = "2222"

				// billing manager
				user3 := models.UserFields{}
				user3.Username = "user3"
				user3.Guid = "3333"

				// auditors
				user4 := models.UserFields{}
				user4.Username = "user4"
				user4.Guid = "4444"

				userRepo.ListUsersByRole = map[string][]models.UserFields{
					models.ORG_MANAGER:     []models.UserFields{user, user2, user3, user4},
					models.BILLING_MANAGER: []models.UserFields{user2, user4},
					models.ORG_AUDITOR:     []models.UserFields{user, user3},
					models.ORG_USER:        []models.UserFields{user, user2, user3, user4},
				}

				requirementsFactory.LoginSuccess = true
				requirementsFactory.Organization = org
				pluginUserModel = []plugin_models.GetOrgUsers_Model{}
				deps.PluginModels.OrgUsers = &pluginUserModel
				updateCommandDependency(true)
			})

			It("populates the plugin model with users with multiple roles", func() {
				runCommand("the-org")

				Ω(pluginUserModel).To(HaveLen(4))
				for _, u := range pluginUserModel {
					switch u.Username {
					case "user1":
						Ω(u.Guid).To(Equal("1111"))
						Ω(u.Roles).To(ConsistOf([]string{models.ORG_MANAGER, models.ORG_AUDITOR}))
						Ω(u.IsAdmin).To(BeTrue())
					case "user2":
						Ω(u.Guid).To(Equal("2222"))
						Ω(u.Roles).To(ConsistOf([]string{models.ORG_MANAGER, models.BILLING_MANAGER}))
					case "user3":
						Ω(u.Guid).To(Equal("3333"))
						Ω(u.Roles).To(ConsistOf([]string{models.ORG_AUDITOR, models.ORG_MANAGER}))
					case "user4":
						Ω(u.Guid).To(Equal("4444"))
						Ω(u.Roles).To(ConsistOf([]string{models.BILLING_MANAGER, models.ORG_MANAGER}))
					default:
						Fail("unexpected user: " + u.Username)
					}
				}

			})

			It("populates the plugin model with users with multiple roles -a flag", func() {
				runCommand("-a", "the-org")

				Ω(pluginUserModel).To(HaveLen(4))
				for _, u := range pluginUserModel {
					switch u.Username {
					case "user1":
						Ω(u.Guid).To(Equal("1111"))
						Ω(u.Roles).To(ConsistOf([]string{models.ORG_USER}))
					case "user2":
						Ω(u.Guid).To(Equal("2222"))
						Ω(u.Roles).To(ConsistOf([]string{models.ORG_USER}))
					case "user3":
						Ω(u.Guid).To(Equal("3333"))
						Ω(u.Roles).To(ConsistOf([]string{models.ORG_USER}))
					case "user4":
						Ω(u.Guid).To(Equal("4444"))
						Ω(u.Roles).To(ConsistOf([]string{models.ORG_USER}))
					default:
						Fail("unexpected user: " + u.Username)
					}
				}

			})

		})

	})
})
