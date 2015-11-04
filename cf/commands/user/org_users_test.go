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
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("org-users", args, requirementsFactory, updateCommandDependency, false)
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

	Context("when logged in and given an org with no users in a particular role", func() {
		var (
			user1, user2 models.UserFields
		)

		BeforeEach(func() {
			org := models.Organization{}
			org.Name = "the-org"
			org.Guid = "the-org-guid"

			user1 = models.UserFields{}
			user1.Username = "user1"
			user2 = models.UserFields{}
			user2.Username = "user2"

			requirementsFactory.LoginSuccess = true
			requirementsFactory.Organization = org
		})

		Context("shows friendly messaage when no users in ORG_MANAGER role", func() {
			It("shows the special users in the given org", func() {
				userRepo.ListUsersInOrgForRoleStub = func(_ string, roleName string) ([]models.UserFields, error) {
					userFields := map[string][]models.UserFields{
						models.ORG_MANAGER:     []models.UserFields{},
						models.BILLING_MANAGER: []models.UserFields{user1},
						models.ORG_AUDITOR:     []models.UserFields{user2},
					}[roleName]
					return userFields, nil
				}

				runCommand("the-org")

				Expect(userRepo.ListUsersInOrgForRoleCallCount()).To(Equal(3))
				for i, expectedRole := range []string{models.ORG_MANAGER, models.BILLING_MANAGER, models.ORG_AUDITOR} {
					orgGuid, actualRole := userRepo.ListUsersInOrgForRoleArgsForCall(i)
					Expect(orgGuid).To(Equal("the-org-guid"))
					Expect(actualRole).To(Equal(expectedRole))
				}

				Expect(ui.Outputs).To(BeInDisplayOrder(
					[]string{"Getting users in org", "the-org", "my-user"},
					[]string{"ORG MANAGER"},
					[]string{"  No ORG MANAGER found"},
					[]string{"BILLING MANAGER"},
					[]string{"  user1"},
					[]string{"ORG AUDITOR"},
					[]string{"  user2"},
				))
			})
		})

		Context("shows friendly messaage when no users in BILLING_MANAGER role", func() {
			It("shows the special users in the given org", func() {
				userRepo.ListUsersInOrgForRoleStub = func(_ string, roleName string) ([]models.UserFields, error) {
					userFields := map[string][]models.UserFields{
						models.ORG_MANAGER:     []models.UserFields{user1},
						models.BILLING_MANAGER: []models.UserFields{},
						models.ORG_AUDITOR:     []models.UserFields{user2},
					}[roleName]
					return userFields, nil
				}

				runCommand("the-org")

				Expect(userRepo.ListUsersInOrgForRoleCallCount()).To(Equal(3))
				for i, expectedRole := range []string{models.ORG_MANAGER, models.BILLING_MANAGER, models.ORG_AUDITOR} {
					orgGuid, actualRole := userRepo.ListUsersInOrgForRoleArgsForCall(i)
					Expect(orgGuid).To(Equal("the-org-guid"))
					Expect(actualRole).To(Equal(expectedRole))
				}

				Expect(ui.Outputs).To(BeInDisplayOrder(
					[]string{"Getting users in org", "the-org", "my-user"},
					[]string{"ORG MANAGER"},
					[]string{"  user1"},
					[]string{"BILLING MANAGER"},
					[]string{"  No BILLING MANAGER found"},
					[]string{"ORG AUDITOR"},
					[]string{"  user2"},
				))
			})
		})

		Context("shows friendly messaage when no users in ORG_AUDITOR role", func() {
			It("shows the special users in the given org", func() {
				userRepo.ListUsersInOrgForRoleStub = func(_ string, roleName string) ([]models.UserFields, error) {
					userFields := map[string][]models.UserFields{
						models.ORG_MANAGER:     []models.UserFields{user1},
						models.BILLING_MANAGER: []models.UserFields{user2},
						models.ORG_AUDITOR:     []models.UserFields{},
					}[roleName]
					return userFields, nil
				}

				runCommand("the-org")

				Expect(userRepo.ListUsersInOrgForRoleCallCount()).To(Equal(3))
				for i, expectedRole := range []string{models.ORG_MANAGER, models.BILLING_MANAGER, models.ORG_AUDITOR} {
					orgGuid, actualRole := userRepo.ListUsersInOrgForRoleArgsForCall(i)
					Expect(orgGuid).To(Equal("the-org-guid"))
					Expect(actualRole).To(Equal(expectedRole))
				}
				Expect(ui.Outputs).To(BeInDisplayOrder(
					[]string{"Getting users in org", "the-org", "my-user"},
					[]string{"ORG MANAGER"},
					[]string{"  user1"},
					[]string{"BILLING MANAGER"},
					[]string{"  user2"},
					[]string{"ORG AUDITOR"},
					[]string{"  No ORG AUDITOR found"},
				))
			})
		})

	})

	Context("when logged in and given an org with users", func() {
		BeforeEach(func() {
			org := models.Organization{}
			org.Name = "the-org"
			org.Guid = "the-org-guid"

			user := models.UserFields{Username: "user1"}
			user2 := models.UserFields{Username: "user2"}
			user3 := models.UserFields{Username: "user3"}
			user4 := models.UserFields{Username: "user4"}
			userRepo.ListUsersInOrgForRoleStub = func(_ string, roleName string) ([]models.UserFields, error) {
				userFields := map[string][]models.UserFields{
					models.ORG_MANAGER:     []models.UserFields{user, user2},
					models.BILLING_MANAGER: []models.UserFields{user4},
					models.ORG_AUDITOR:     []models.UserFields{user3},
				}[roleName]
				return userFields, nil
			}

			requirementsFactory.LoginSuccess = true
			requirementsFactory.Organization = org
		})

		It("shows the special users in the given org", func() {
			runCommand("the-org")

			orgGuid, _ := userRepo.ListUsersInOrgForRoleArgsForCall(0)
			Expect(orgGuid).To(Equal("the-org-guid"))
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
				user := models.UserFields{Username: "user1"}
				user2 := models.UserFields{Username: "user2"}
				userRepo.ListUsersInOrgForRoleStub = func(_ string, roleName string) ([]models.UserFields, error) {
					userFields := map[string][]models.UserFields{
						models.ORG_USER: []models.UserFields{user, user2},
					}[roleName]
					return userFields, nil
				}
			})

			It("lists all org users, regardless of role", func() {
				runCommand("-a", "the-org")

				orgGuid, _ := userRepo.ListUsersInOrgForRoleArgsForCall(0)
				Expect(orgGuid).To(Equal("the-org-guid"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting users in org", "the-org", "my-user"},
					[]string{"USERS"},
					[]string{"user1"},
					[]string{"user2"},
				))
			})
		})

		Context("when cc api verson is >= 2.21.0", func() {
			It("calls ListUsersInOrgForRoleWithNoUAA()", func() {
				configRepo.SetApiVersion("2.22.0")
				runCommand("the-org")

				Expect(userRepo.ListUsersInOrgForRoleWithNoUAACallCount()).To(BeNumerically(">=", 1))
				Expect(userRepo.ListUsersInOrgForRoleCallCount()).To(Equal(0))
			})
		})

		Context("when cc api verson is < 2.21.0", func() {
			It("calls ListUsersInOrgForRole()", func() {
				configRepo.SetApiVersion("2.20.0")
				runCommand("the-org")

				Expect(userRepo.ListUsersInOrgForRoleWithNoUAACallCount()).To(Equal(0))
				Expect(userRepo.ListUsersInOrgForRoleCallCount()).To(BeNumerically(">=", 1))
			})
		})
	})

	Describe("when invoked by a plugin", func() {
		var (
			pluginUserModel []plugin_models.GetOrgUsers_Model
		)

		BeforeEach(func() {
			configRepo.SetApiVersion("2.22.0")
		})

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

				userRepo.ListUsersInOrgForRoleWithNoUAAStub = func(_ string, roleName string) ([]models.UserFields, error) {
					userFields := map[string][]models.UserFields{
						models.ORG_MANAGER:     []models.UserFields{user, user2},
						models.BILLING_MANAGER: []models.UserFields{user4},
						models.ORG_AUDITOR:     []models.UserFields{user3},
						models.ORG_USER:        []models.UserFields{user3},
					}[roleName]
					return userFields, nil
				}

				requirementsFactory.LoginSuccess = true
				requirementsFactory.Organization = org
				pluginUserModel = []plugin_models.GetOrgUsers_Model{}
				deps.PluginModels.OrgUsers = &pluginUserModel
			})

			It("populates the plugin model with users with single roles", func() {
				testcmd.RunCliCommand("org-users", []string{"the-org"}, requirementsFactory, updateCommandDependency, true)
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
				testcmd.RunCliCommand("org-users", []string{"-a", "the-org"}, requirementsFactory, updateCommandDependency, true)
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

				userRepo.ListUsersInOrgForRoleWithNoUAAStub = func(_ string, roleName string) ([]models.UserFields, error) {
					userFields := map[string][]models.UserFields{
						models.ORG_MANAGER:     []models.UserFields{user, user2, user3, user4},
						models.BILLING_MANAGER: []models.UserFields{user2, user4},
						models.ORG_AUDITOR:     []models.UserFields{user, user3},
						models.ORG_USER:        []models.UserFields{user, user2, user3, user4},
					}[roleName]
					return userFields, nil
				}

				requirementsFactory.LoginSuccess = true
				requirementsFactory.Organization = org
				pluginUserModel = []plugin_models.GetOrgUsers_Model{}
				deps.PluginModels.OrgUsers = &pluginUserModel
			})

			It("populates the plugin model with users with multiple roles", func() {
				testcmd.RunCliCommand("org-users", []string{"the-org"}, requirementsFactory, updateCommandDependency, true)

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
				testcmd.RunCliCommand("org-users", []string{"-a", "the-org"}, requirementsFactory, updateCommandDependency, true)

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
