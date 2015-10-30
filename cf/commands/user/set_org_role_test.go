package user_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakeflag "github.com/cloudfoundry/cli/cf/api/feature_flags/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
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

var _ = Describe("set-org-role command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		userRepo            *testapi.FakeUserRepository
		configRepo          core_config.Repository
		flagRepo            *fakeflag.FakeFeatureFlagRepository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.
			SetUserRepository(userRepo).
			SetFeatureFlagRepository(flagRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("set-org-role").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()

		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		userRepo = &testapi.FakeUserRepository{}
		flagRepo = &fakeflag.FakeFeatureFlagRepository{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("set-org-role", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			Expect(runCommand("hey", "there", "jude")).To(BeFalse())
		})

		It("fails with usage when not provided exactly three args", func() {
			runCommand("one fish", "two fish") // red fish, blue fish
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true

			org := models.Organization{}
			org.Guid = "my-org-guid"
			org.Name = "my-org"
			requirementsFactory.UserFields = models.UserFields{Guid: "my-user-guid", Username: "my-user"}
			requirementsFactory.Organization = org
		})

		Context("and the API supports role-based auth management for org/space", func() {
			BeforeEach(func() {
				configRepo.SetApiVersion("2.37.0")
			})

			It("retrieves the feature flag determining which endpoints to use", func() {
				runCommand("some-user", "some-org", "OrgManager")
				Expect(flagRepo.FindByNameArgsForCall(0)).To(Equal("set_roles_by_username"))
			})

			Context("and the feature flag is up", func() {
				BeforeEach(func() {
					flag := models.FeatureFlag{
						Name:    "set_roles_by_username",
						Enabled: true,
					}
					flagRepo.FindByNameReturns(flag, nil)
				})

				It("sets the role using a username", func() {
					runCommand("irrelevant-user", "some-org", "OrgManager")
					Expect(userRepo.SetOrgRoleUsername).To(Equal("my-user"))
				})

				It("copes with user repo failure", func() {
					userRepo.SetOrgRoleByUsernameError = errors.New("TROUBLE AT CC")
					runCommand("some-user", "some-org", "OrgManager")
					Expect(ui.Outputs).To(BeInDisplayOrder(
						[]string{"FAILED"},
						[]string{"TROUBLE AT CC"},
					))
				})
			})

			Context("but the feature flag is down", func() {
				It("sets the role using a GUID", func() {
					runCommand("some-user", "some-org", "OrgManager")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Assigning role", "OrgManager", "my-user", "my-org", "my-user"},
						[]string{"OK"},
					))
					Expect(userRepo.SetOrgRoleCalled).To(BeTrue())
				})
			})

			It("copes with feature flag failure", func() {
				flagRepo.FindByNameReturns(models.FeatureFlag{}, errors.New("NO ONE AT FLAG SHOP"))
				runCommand("some-user", "some-org", "OrgManager")
				Expect(ui.Outputs).To(BeInDisplayOrder(
					[]string{"FAILED"},
					[]string{"NO ONE AT FLAG SHOP"},
				))
			})
		})

		Context("when API doesn't support role-based auth management for org/space", func() {
			BeforeEach(func() {
				configRepo.SetApiVersion("2.36.939393")
				flag := models.FeatureFlag{
					Name:    "set_roles_by_username",
					Enabled: true,
				}
				flagRepo.FindByNameReturns(flag, nil)
			})

			It("sets the role using a GUID", func() {
				runCommand("some-user", "some-org", "OrgManager")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Assigning role", "OrgManager", "my-user", "my-org", "my-user"},
					[]string{"OK"},
				))
				Expect(userRepo.SetOrgRoleCalled).To(BeTrue())
			})
		})
	})
})
