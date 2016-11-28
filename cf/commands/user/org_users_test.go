package user_test

import (
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	"code.cloudfoundry.org/cli/plugin/models"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("org-users command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          coreconfig.Repository
		userRepo            *apifakes.FakeUserRepository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetUserRepository(userRepo)

		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("org-users").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		userRepo = new(apifakes.FakeUserRepository)
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		deps = commandregistry.NewDependency(os.Stdout, new(tracefakes.FakePrinter), "")
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("org-users", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails with usage when invoked without an org name", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
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
			org.GUID = "the-org-guid"

			user1 = models.UserFields{}
			user1.Username = "user1"
			user2 = models.UserFields{}
			user2.Username = "user2"

			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			organizationReq := new(requirementsfakes.FakeOrganizationRequirement)
			organizationReq.GetOrganizationReturns(org)
			requirementsFactory.NewOrganizationRequirementReturns(organizationReq)
		})

		Context("shows friendly messaage when no users in ORG_MANAGER role", func() {
			It("shows the special users in the given org", func() {
				userRepo.ListUsersInOrgForRoleStub = func(_ string, roleName models.Role) ([]models.UserFields, error) {
					userFields := map[models.Role][]models.UserFields{
						models.RoleOrgManager:     {},
						models.RoleBillingManager: {user1},
						models.RoleOrgAuditor:     {user2},
					}[roleName]
					return userFields, nil
				}

				runCommand("the-org")

				Expect(userRepo.ListUsersInOrgForRoleCallCount()).To(Equal(3))
				for i, expectedRole := range []models.Role{models.RoleOrgManager, models.RoleBillingManager, models.RoleOrgAuditor} {
					orgGUID, actualRole := userRepo.ListUsersInOrgForRoleArgsForCall(i)
					Expect(orgGUID).To(Equal("the-org-guid"))
					Expect(actualRole).To(Equal(expectedRole))
				}

				Expect(ui.Outputs()).To(BeInDisplayOrder(
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
				userRepo.ListUsersInOrgForRoleStub = func(_ string, roleName models.Role) ([]models.UserFields, error) {
					userFields := map[models.Role][]models.UserFields{
						models.RoleOrgManager:     {user1},
						models.RoleBillingManager: {},
						models.RoleOrgAuditor:     {user2},
					}[roleName]
					return userFields, nil
				}

				runCommand("the-org")

				Expect(userRepo.ListUsersInOrgForRoleCallCount()).To(Equal(3))
				for i, expectedRole := range []models.Role{models.RoleOrgManager, models.RoleBillingManager, models.RoleOrgAuditor} {
					orgGUID, actualRole := userRepo.ListUsersInOrgForRoleArgsForCall(i)
					Expect(orgGUID).To(Equal("the-org-guid"))
					Expect(actualRole).To(Equal(expectedRole))
				}

				Expect(ui.Outputs()).To(BeInDisplayOrder(
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
				userRepo.ListUsersInOrgForRoleStub = func(_ string, roleName models.Role) ([]models.UserFields, error) {
					userFields := map[models.Role][]models.UserFields{
						models.RoleOrgManager:     {user1},
						models.RoleBillingManager: {user2},
						models.RoleOrgAuditor:     {},
					}[roleName]
					return userFields, nil
				}

				runCommand("the-org")

				Expect(userRepo.ListUsersInOrgForRoleCallCount()).To(Equal(3))
				for i, expectedRole := range []models.Role{models.RoleOrgManager, models.RoleBillingManager, models.RoleOrgAuditor} {
					orgGUID, actualRole := userRepo.ListUsersInOrgForRoleArgsForCall(i)
					Expect(orgGUID).To(Equal("the-org-guid"))
					Expect(actualRole).To(Equal(expectedRole))
				}
				Expect(ui.Outputs()).To(BeInDisplayOrder(
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
			org.GUID = "the-org-guid"

			user := models.UserFields{Username: "user1"}
			user2 := models.UserFields{Username: "user2"}
			user3 := models.UserFields{Username: "user3"}
			user4 := models.UserFields{Username: "user4"}
			userRepo.ListUsersInOrgForRoleStub = func(_ string, roleName models.Role) ([]models.UserFields, error) {
				userFields := map[models.Role][]models.UserFields{
					models.RoleOrgManager:     {user, user2},
					models.RoleBillingManager: {user4},
					models.RoleOrgAuditor:     {user3},
				}[roleName]
				return userFields, nil
			}

			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			organizationReq := new(requirementsfakes.FakeOrganizationRequirement)
			organizationReq.GetOrganizationReturns(org)
			requirementsFactory.NewOrganizationRequirementReturns(organizationReq)
		})

		It("shows the special users in the given org", func() {
			runCommand("the-org")

			orgGUID, _ := userRepo.ListUsersInOrgForRoleArgsForCall(0)
			Expect(orgGUID).To(Equal("the-org-guid"))
			Expect(ui.Outputs()).To(ContainSubstrings(
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
				userRepo.ListUsersInOrgForRoleStub = func(_ string, roleName models.Role) ([]models.UserFields, error) {
					userFields := map[models.Role][]models.UserFields{
						models.RoleOrgUser: {user, user2},
					}[roleName]
					return userFields, nil
				}
			})

			It("lists all org users, regardless of role", func() {
				runCommand("-a", "the-org")

				orgGUID, _ := userRepo.ListUsersInOrgForRoleArgsForCall(0)
				Expect(orgGUID).To(Equal("the-org-guid"))
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting users in org", "the-org", "my-user"},
					[]string{"USERS"},
					[]string{"user1"},
					[]string{"user2"},
				))
			})
		})

		Context("when cc api verson is >= 2.21.0", func() {
			It("calls ListUsersInOrgForRoleWithNoUAA()", func() {
				configRepo.SetAPIVersion("2.22.0")
				runCommand("the-org")

				Expect(userRepo.ListUsersInOrgForRoleWithNoUAACallCount()).To(BeNumerically(">=", 1))
				Expect(userRepo.ListUsersInOrgForRoleCallCount()).To(Equal(0))
			})
		})

		Context("when cc api verson is < 2.21.0", func() {
			It("calls ListUsersInOrgForRole()", func() {
				configRepo.SetAPIVersion("2.20.0")
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
			configRepo.SetAPIVersion("2.22.0")
		})

		Context("single roles", func() {

			BeforeEach(func() {
				org := models.Organization{}
				org.Name = "the-org"
				org.GUID = "the-org-guid"

				// org managers
				user := models.UserFields{}
				user.Username = "user1"
				user.GUID = "1111"

				user2 := models.UserFields{}
				user2.Username = "user2"
				user2.GUID = "2222"

				// billing manager
				user3 := models.UserFields{}
				user3.Username = "user3"
				user3.GUID = "3333"

				// auditors
				user4 := models.UserFields{}
				user4.Username = "user4"
				user4.GUID = "4444"

				userRepo.ListUsersInOrgForRoleWithNoUAAStub = func(_ string, roleName models.Role) ([]models.UserFields, error) {
					userFields := map[models.Role][]models.UserFields{
						models.RoleOrgManager:     {user, user2},
						models.RoleBillingManager: {user4},
						models.RoleOrgAuditor:     {user3},
						models.RoleOrgUser:        {user3},
					}[roleName]
					return userFields, nil
				}

				requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
				organizationReq := new(requirementsfakes.FakeOrganizationRequirement)
				organizationReq.GetOrganizationReturns(org)
				requirementsFactory.NewOrganizationRequirementReturns(organizationReq)
				pluginUserModel = []plugin_models.GetOrgUsers_Model{}
				deps.PluginModels.OrgUsers = &pluginUserModel
			})

			It("populates the plugin model with users with single roles", func() {
				testcmd.RunCLICommand("org-users", []string{"the-org"}, requirementsFactory, updateCommandDependency, true, ui)
				Expect(pluginUserModel).To(HaveLen(4))

				for _, u := range pluginUserModel {
					switch u.Username {
					case "user1":
						Expect(u.Guid).To(Equal("1111"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleOrgManager"}))
					case "user2":
						Expect(u.Guid).To(Equal("2222"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleOrgManager"}))
					case "user3":
						Expect(u.Guid).To(Equal("3333"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleOrgAuditor"}))
					case "user4":
						Expect(u.Guid).To(Equal("4444"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleBillingManager"}))
					default:
						Fail("unexpected user: " + u.Username)
					}
				}

			})

			It("populates the plugin model with users with single roles -a flag", func() {
				testcmd.RunCLICommand("org-users", []string{"-a", "the-org"}, requirementsFactory, updateCommandDependency, true, ui)
				Expect(pluginUserModel).To(HaveLen(1))
				Expect(pluginUserModel[0].Username).To(Equal("user3"))
				Expect(pluginUserModel[0].Guid).To(Equal("3333"))
				Expect(pluginUserModel[0].Roles[0]).To(Equal("RoleOrgUser"))
			})

		})

		Context("multiple roles", func() {

			BeforeEach(func() {
				org := models.Organization{}
				org.Name = "the-org"
				org.GUID = "the-org-guid"

				// org managers
				user := models.UserFields{}
				user.Username = "user1"
				user.GUID = "1111"
				user.IsAdmin = true

				user2 := models.UserFields{}
				user2.Username = "user2"
				user2.GUID = "2222"

				// billing manager
				user3 := models.UserFields{}
				user3.Username = "user3"
				user3.GUID = "3333"

				// auditors
				user4 := models.UserFields{}
				user4.Username = "user4"
				user4.GUID = "4444"

				userRepo.ListUsersInOrgForRoleWithNoUAAStub = func(_ string, roleName models.Role) ([]models.UserFields, error) {
					userFields := map[models.Role][]models.UserFields{
						models.RoleOrgManager:     {user, user2, user3, user4},
						models.RoleBillingManager: {user2, user4},
						models.RoleOrgAuditor:     {user, user3},
						models.RoleOrgUser:        {user, user2, user3, user4},
					}[roleName]
					return userFields, nil
				}

				requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
				organizationReq := new(requirementsfakes.FakeOrganizationRequirement)
				organizationReq.GetOrganizationReturns(org)
				requirementsFactory.NewOrganizationRequirementReturns(organizationReq)
				pluginUserModel = []plugin_models.GetOrgUsers_Model{}
				deps.PluginModels.OrgUsers = &pluginUserModel
			})

			It("populates the plugin model with users with multiple roles", func() {
				testcmd.RunCLICommand("org-users", []string{"the-org"}, requirementsFactory, updateCommandDependency, true, ui)

				Expect(pluginUserModel).To(HaveLen(4))
				for _, u := range pluginUserModel {
					switch u.Username {
					case "user1":
						Expect(u.Guid).To(Equal("1111"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleOrgManager", "RoleOrgAuditor"}))
						Expect(u.IsAdmin).To(BeTrue())
					case "user2":
						Expect(u.Guid).To(Equal("2222"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleOrgManager", "RoleBillingManager"}))
					case "user3":
						Expect(u.Guid).To(Equal("3333"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleOrgAuditor", "RoleOrgManager"}))
					case "user4":
						Expect(u.Guid).To(Equal("4444"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleBillingManager", "RoleOrgManager"}))
					default:
						Fail("unexpected user: " + u.Username)
					}
				}

			})

			It("populates the plugin model with users with multiple roles -a flag", func() {
				testcmd.RunCLICommand("org-users", []string{"-a", "the-org"}, requirementsFactory, updateCommandDependency, true, ui)

				Expect(pluginUserModel).To(HaveLen(4))
				for _, u := range pluginUserModel {
					switch u.Username {
					case "user1":
						Expect(u.Guid).To(Equal("1111"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleOrgUser"}))
					case "user2":
						Expect(u.Guid).To(Equal("2222"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleOrgUser"}))
					case "user3":
						Expect(u.Guid).To(Equal("3333"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleOrgUser"}))
					case "user4":
						Expect(u.Guid).To(Equal("4444"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleOrgUser"}))
					default:
						Fail("unexpected user: " + u.Username)
					}
				}

			})

		})

	})
})
