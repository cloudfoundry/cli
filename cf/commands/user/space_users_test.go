package user_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/api/spaces/spacesfakes"
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

var _ = Describe("space-users command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		spaceRepo           *spacesfakes.FakeSpaceRepository
		userRepo            *apifakes.FakeUserRepository
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetUserRepository(userRepo)
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)

		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("space-users").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		ui = &testterm.FakeUI{}
		requirementsFactory = new(requirementsfakes.FakeFactory)
		spaceRepo = new(spacesfakes.FakeSpaceRepository)
		userRepo = new(apifakes.FakeUserRepository)
		deps = commandregistry.NewDependency(os.Stdout, new(tracefakes.FakePrinter), "")
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("space-users", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-org", "my-space")).To(BeFalse())
		})

		It("succeeds when logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			organizationReq := new(requirementsfakes.FakeOrganizationRequirement)
			organizationReq.GetOrganizationReturns(
				models.Organization{
					OrganizationFields: models.OrganizationFields{
						Name: "some-org",
					},
				},
			)
			spaceRepo.FindByNameInOrgReturns(
				models.Space{
					SpaceFields: models.SpaceFields{
						Name: "whatever-space",
					},
				}, nil)
			requirementsFactory.NewOrganizationRequirementReturns(organizationReq)
			passed := runCommand("some-org", "whatever-space")

			Expect(passed).To(BeTrue())
			Expect(ui.Outputs()).To(ContainSubstrings([]string{"Getting users in org some-org / space whatever-space as my-user"}))
		})
	})

	It("fails with usage when not invoked with exactly two args", func() {
		runCommand("my-org")
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires arguments"},
		))
	})

	Context("when logged in and given some users in the org and space", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})

			org := models.Organization{}
			org.Name = "Org1"
			org.GUID = "org1-guid"
			space := models.Space{}
			space.Name = "Space1"
			space.GUID = "space1-guid"

			organizationReq := new(requirementsfakes.FakeOrganizationRequirement)
			organizationReq.GetOrganizationReturns(org)
			requirementsFactory.NewOrganizationRequirementReturns(organizationReq)
			spaceRepo.FindByNameInOrgReturns(space, nil)

			user := models.UserFields{}
			user.Username = "user1"
			user2 := models.UserFields{}
			user2.Username = "user2"
			user3 := models.UserFields{}
			user3.Username = "user3"
			user4 := models.UserFields{}
			user4.Username = "user4"
			userRepo.ListUsersInSpaceForRoleStub = func(_ string, roleName models.Role) ([]models.UserFields, error) {
				userFields := map[models.Role][]models.UserFields{
					models.RoleSpaceManager:   {user, user2},
					models.RoleSpaceDeveloper: {user4},
					models.RoleSpaceAuditor:   {user3},
				}[roleName]
				return userFields, nil
			}
		})

		It("tells you about the space users in the given space", func() {
			runCommand("my-org", "my-space")

			actualSpaceName, actualOrgGUID := spaceRepo.FindByNameInOrgArgsForCall(0)
			Expect(actualSpaceName).To(Equal("my-space"))
			Expect(actualOrgGUID).To(Equal("org1-guid"))

			Expect(userRepo.ListUsersInSpaceForRoleCallCount()).To(Equal(3))
			for i, expectedRole := range []models.Role{models.RoleSpaceManager, models.RoleSpaceDeveloper, models.RoleSpaceAuditor} {
				spaceGUID, actualRole := userRepo.ListUsersInSpaceForRoleArgsForCall(i)
				Expect(spaceGUID).To(Equal("space1-guid"))
				Expect(actualRole).To(Equal(expectedRole))
			}

			Expect(ui.Outputs()).To(BeInDisplayOrder(
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

		Context("when cc api verson is >= 2.21.0", func() {
			BeforeEach(func() {
				configRepo.SetAPIVersion("2.22.0")
			})

			It("calls ListUsersInSpaceForRoleWithNoUAA()", func() {
				runCommand("my-org", "my-sapce")

				Expect(userRepo.ListUsersInSpaceForRoleWithNoUAACallCount()).To(BeNumerically(">=", 1))
				Expect(userRepo.ListUsersInSpaceForRoleCallCount()).To(Equal(0))
			})

			It("fails with an error when user network call fails", func() {
				userRepo.ListUsersInSpaceForRoleWithNoUAAStub = func(_ string, role models.Role) ([]models.UserFields, error) {
					if role == models.RoleSpaceManager {
						return []models.UserFields{}, errors.New("internet badness occurred")
					}
					return []models.UserFields{}, nil
				}
				runCommand("my-org", "my-space")
				Expect(ui.Outputs()).To(BeInDisplayOrder(
					[]string{"Getting users in org", "Org1"},
					[]string{"internet badness occurred"},
				))
			})
		})

		Context("when cc api verson is < 2.21.0", func() {
			It("calls ListUsersInSpaceForRole()", func() {
				configRepo.SetAPIVersion("2.20.0")
				runCommand("my-org", "my-space")

				Expect(userRepo.ListUsersInSpaceForRoleWithNoUAACallCount()).To(Equal(0))
				Expect(userRepo.ListUsersInSpaceForRoleCallCount()).To(BeNumerically(">=", 1))
			})
		})
	})

	Context("when logged in and there are no non-managers in the space", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})

			org := models.Organization{}
			org.Name = "Org1"
			org.GUID = "org1-guid"
			space := models.Space{}
			space.Name = "Space1"
			space.GUID = "space1-guid"

			organizationReq := new(requirementsfakes.FakeOrganizationRequirement)
			organizationReq.GetOrganizationReturns(org)
			requirementsFactory.NewOrganizationRequirementReturns(organizationReq)
			spaceRepo.FindByNameInOrgReturns(space, nil)

			user := models.UserFields{}
			user.Username = "mr-pointy-hair"
			userRepo.ListUsersInSpaceForRoleStub = func(_ string, roleName models.Role) ([]models.UserFields, error) {
				userFields := map[models.Role][]models.UserFields{
					models.RoleSpaceManager:   {user},
					models.RoleSpaceDeveloper: {},
					models.RoleSpaceAuditor:   {},
				}[roleName]
				return userFields, nil
			}
		})

		It("shows a friendly message when there are no users in a role", func() {
			runCommand("my-org", "my-space")

			Expect(ui.Outputs()).To(BeInDisplayOrder(
				[]string{"Getting users in org"},
				[]string{"SPACE MANAGER"},
				[]string{"mr-pointy-hair"},
				[]string{"SPACE DEVELOPER"},
				[]string{"No SPACE DEVELOPER found"},
				[]string{"SPACE AUDITOR"},
				[]string{"No SPACE AUDITOR found"},
			))
		})
	})

	Describe("when invoked by a plugin", func() {
		var (
			pluginUserModel []plugin_models.GetSpaceUsers_Model
		)

		BeforeEach(func() {
			configRepo.SetAPIVersion("2.22.0")
		})

		Context("single roles", func() {
			BeforeEach(func() {
				org := models.Organization{}
				org.Name = "the-org"
				org.GUID = "the-org-guid"

				space := models.Space{}
				space.Name = "the-space"
				space.GUID = "the-space-guid"

				// space managers
				user := models.UserFields{}
				user.Username = "user1"
				user.GUID = "1111"

				user2 := models.UserFields{}
				user2.Username = "user2"
				user2.GUID = "2222"

				// space auditor
				user3 := models.UserFields{}
				user3.Username = "user3"
				user3.GUID = "3333"

				// space developer
				user4 := models.UserFields{}
				user4.Username = "user4"
				user4.GUID = "4444"

				userRepo.ListUsersInSpaceForRoleWithNoUAAStub = func(_ string, roleName models.Role) ([]models.UserFields, error) {
					userFields := map[models.Role][]models.UserFields{
						models.RoleSpaceManager:   {user, user2},
						models.RoleSpaceDeveloper: {user4},
						models.RoleSpaceAuditor:   {user3},
					}[roleName]
					return userFields, nil
				}

				requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
				organizationReq := new(requirementsfakes.FakeOrganizationRequirement)
				organizationReq.GetOrganizationReturns(org)
				requirementsFactory.NewOrganizationRequirementReturns(organizationReq)
				pluginUserModel = []plugin_models.GetSpaceUsers_Model{}
				deps.PluginModels.SpaceUsers = &pluginUserModel
			})

			It("populates the plugin model with users with single roles", func() {
				testcmd.RunCLICommand("space-users", []string{"the-org", "the-space"}, requirementsFactory, updateCommandDependency, true, ui)

				Expect(pluginUserModel).To(HaveLen(4))
				for _, u := range pluginUserModel {
					switch u.Username {
					case "user1":
						Expect(u.Guid).To(Equal("1111"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleSpaceManager"}))
					case "user2":
						Expect(u.Guid).To(Equal("2222"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleSpaceManager"}))
					case "user3":
						Expect(u.Guid).To(Equal("3333"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleSpaceAuditor"}))
					case "user4":
						Expect(u.Guid).To(Equal("4444"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleSpaceDeveloper"}))
					default:
						Fail("unexpected user: " + u.Username)
					}
				}
			})
		})

		Context("multiple roles", func() {
			BeforeEach(func() {
				org := models.Organization{}
				org.Name = "the-org"
				org.GUID = "the-org-guid"

				space := models.Space{}
				space.Name = "the-space"
				space.GUID = "the-space-guid"

				// space managers
				user := models.UserFields{}
				user.Username = "user1"
				user.GUID = "1111"

				user2 := models.UserFields{}
				user2.Username = "user2"
				user2.GUID = "2222"

				// space auditor
				user3 := models.UserFields{}
				user3.Username = "user3"
				user3.GUID = "3333"

				// space developer
				user4 := models.UserFields{}
				user4.Username = "user4"
				user4.GUID = "4444"

				userRepo.ListUsersInSpaceForRoleWithNoUAAStub = func(_ string, roleName models.Role) ([]models.UserFields, error) {
					userFields := map[models.Role][]models.UserFields{
						models.RoleSpaceManager:   {user, user2, user3, user4},
						models.RoleSpaceDeveloper: {user2, user4},
						models.RoleSpaceAuditor:   {user, user3},
					}[roleName]
					return userFields, nil
				}

				requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
				organizationReq := new(requirementsfakes.FakeOrganizationRequirement)
				organizationReq.GetOrganizationReturns(org)
				requirementsFactory.NewOrganizationRequirementReturns(organizationReq)
				pluginUserModel = []plugin_models.GetSpaceUsers_Model{}
				deps.PluginModels.SpaceUsers = &pluginUserModel
			})

			It("populates the plugin model with users with multiple roles", func() {
				testcmd.RunCLICommand("space-users", []string{"the-org", "the-space"}, requirementsFactory, updateCommandDependency, true, ui)

				Expect(pluginUserModel).To(HaveLen(4))
				for _, u := range pluginUserModel {
					switch u.Username {
					case "user1":
						Expect(u.Guid).To(Equal("1111"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleSpaceManager", "RoleSpaceAuditor"}))
					case "user2":
						Expect(u.Guid).To(Equal("2222"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleSpaceManager", "RoleSpaceDeveloper"}))
					case "user3":
						Expect(u.Guid).To(Equal("3333"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleSpaceManager", "RoleSpaceAuditor"}))
					case "user4":
						Expect(u.Guid).To(Equal("4444"))
						Expect(u.Roles).To(ConsistOf([]string{"RoleSpaceManager", "RoleSpaceDeveloper"}))
					default:
						Fail("unexpected user: " + u.Username)
					}
				}
			})
		})
	})
})
