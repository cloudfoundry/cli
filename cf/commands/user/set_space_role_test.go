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

var _ = Describe("set-space-role command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		spaceRepo           *testapi.FakeSpaceRepository
		userRepo            *testapi.FakeUserRepository
		configRepo          core_config.Repository
		flagRepo            *fakeflag.FakeFeatureFlagRepository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		deps.RepoLocator = deps.RepoLocator.SetUserRepository(userRepo)
		deps.RepoLocator = deps.RepoLocator.SetFeatureFlagRepository(flagRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("set-space-role").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		accessToken, err := testconfig.EncodeAccessToken(core_config.TokenInfo{Username: "current-user"})
		Expect(err).NotTo(HaveOccurred())
		configRepo.SetAccessToken(accessToken)

		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		spaceRepo = &testapi.FakeSpaceRepository{}
		userRepo = &testapi.FakeUserRepository{}
		flagRepo = &fakeflag.FakeFeatureFlagRepository{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("set-space-role", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			Expect(runCommand("username", "org", "space", "role")).To(BeFalse())
		})

		It("succeeds when logged in", func() {
			requirementsFactory.LoginSuccess = true
			passed := runCommand("username", "org", "space", "role")

			Expect(passed).To(BeTrue())
			Expect(requirementsFactory.UserUsername).To(Equal("username"))
			Expect(requirementsFactory.OrganizationName).To(Equal("org"))
		})
	})

	Context("when requirements are satisfied", func() {
		It("fails with usage when not provided exactly four args", func() {
			runCommand("foo", "bar", "baz")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("does not fail with usage when provided four args", func() {
			runCommand("whatever", "these", "are", "args")
			Expect(ui.Outputs).ToNot(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		Context("setting space role", func() {
			BeforeEach(func() {
				requirementsFactory.LoginSuccess = true

				org := models.Organization{}
				org.Guid = "my-org-guid"
				org.Name = "my-org"

				requirementsFactory.UserFields = models.UserFields{Guid: "my-user-guid", Username: "my-user"}
				requirementsFactory.Organization = org

				spaceRepo.FindByNameInOrgSpace = models.Space{}
				spaceRepo.FindByNameInOrgSpace.Guid = "my-space-guid"
				spaceRepo.FindByNameInOrgSpace.Name = "my-space"
				spaceRepo.FindByNameInOrgSpace.Organization = org.OrganizationFields
			})

			Context("when CC version allows space role set by username", func() {
				BeforeEach(func() {
					configRepo.SetApiVersion("2.37.0")
				})

				Context("when retriving feature flag 'set_roles_by_username' returns an error", func() {
					BeforeEach(func() {
						flagRepo.FindByNameReturns(models.FeatureFlag{}, errors.New("something broke"))
					})

					It("returns the error", func() {
						runCommand("some-user", "some-org", "some-space", "SpaceManager")

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"something broke"},
						))
					})
				})

				Context("when feature flag 'set_roles_by_username' is enabled", func() {
					BeforeEach(func() {
						flagRepo.FindByNameReturns(models.FeatureFlag{Enabled: true}, nil)
					})

					Context("when setting role succeed", func() {
						It("uses the new endpoint to set space role by name", func() {
							runCommand("my-user", "some-org", "some-space", "SpaceManager")

							Expect(userRepo.SetSpaceRoleCalled).To(BeFalse())
							Expect(userRepo.SetSpaceRoleByUsernameCalled).To(BeTrue())
						})
					})

					Context("when setting role fails", func() {
						It("returns the error to user", func() {
							userRepo.SetSpaceRoleByUsernameError = errors.New("oh no, it is broken")

							runCommand("my-user", "some-org", "some-space", "SpaceManager")

							Expect(ui.Outputs).To(ContainSubstrings(
								[]string{"FAILED"},
								[]string{"it is broken"},
							))

						})
					})
				})

				Context("when feature flag 'set_roles_by_username' is disabled", func() {
					BeforeEach(func() {
						flagRepo.FindByNameReturns(models.FeatureFlag{Enabled: false}, nil)
					})

					It("uses the old endpoint to set space role by user guid", func() {
						runCommand("my-user", "some-org", "some-space", "SpaceManager")

						Expect(userRepo.SetSpaceRoleCalled).To(BeTrue())
						Expect(userRepo.SetSpaceRoleByUsernameCalled).To(BeFalse())
					})
				})
			})

			Context("when CC version only support space role set by user guid", func() {
				BeforeEach(func() {
					configRepo.SetApiVersion("2.36.9")
				})

				It("sets the given space role on the given user by user guid", func() {
					runCommand("some-user", "some-org", "some-space", "SpaceManager")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Assigning role ", "SpaceManager", "my-user", "my-org", "my-space", "current-user"},
						[]string{"OK"},
					))

					Expect(spaceRepo.FindByNameInOrgName).To(Equal("some-space"))
					Expect(spaceRepo.FindByNameInOrgOrgGuid).To(Equal("my-org-guid"))

					Expect(userRepo.SetSpaceRoleCalled).To(BeTrue())
				})
			})
		})

	})
})
