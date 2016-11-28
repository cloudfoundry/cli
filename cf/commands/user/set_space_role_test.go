package user_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/user"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/api/featureflags/featureflagsfakes"
	"code.cloudfoundry.org/cli/cf/api/spaces/spacesfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetSpaceRole", func() {
	var (
		ui         *testterm.FakeUI
		configRepo coreconfig.Repository
		userRepo   *apifakes.FakeUserRepository
		spaceRepo  *spacesfakes.FakeSpaceRepository
		flagRepo   *featureflagsfakes.FakeFeatureFlagRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement        requirements.Requirement
		userRequirement         *requirementsfakes.FakeUserRequirement
		organizationRequirement *requirementsfakes.FakeOrganizationRequirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		userRepo = new(apifakes.FakeUserRepository)
		repoLocator := deps.RepoLocator.SetUserRepository(userRepo)
		spaceRepo = new(spacesfakes.FakeSpaceRepository)
		repoLocator = repoLocator.SetSpaceRepository(spaceRepo)
		flagRepo = new(featureflagsfakes.FakeFeatureFlagRepository)
		repoLocator = repoLocator.SetFeatureFlagRepository(flagRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &user.SetSpaceRole{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(map[string]flags.FlagSet{})

		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{}
		factory.NewLoginRequirementReturns(loginRequirement)

		userRequirement = new(requirementsfakes.FakeUserRequirement)
		userRequirement.ExecuteReturns(nil)
		factory.NewUserRequirementReturns(userRequirement)

		organizationRequirement = new(requirementsfakes.FakeOrganizationRequirement)
		organizationRequirement.ExecuteReturns(nil)
		factory.NewOrganizationRequirementReturns(organizationRequirement)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly four args", func() {
			BeforeEach(func() {
				flagContext.Parse("the-user-name", "the-org-name", "the-space-name")
			})

			It("fails with usage", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Incorrect Usage. Requires USERNAME, ORG, SPACE, ROLE as arguments"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})
		})

		Context("when provided four args", func() {
			BeforeEach(func() {
				flagContext.Parse("the-user-name", "the-org-name", "the-space-name", "SpaceManager")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))

				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns an OrgRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewOrganizationRequirementCallCount()).To(Equal(1))
				Expect(factory.NewOrganizationRequirementArgsForCall(0)).To(Equal("the-org-name"))

				Expect(actualRequirements).To(ContainElement(organizationRequirement))
			})

			Context("when the config version is >=2.37.0", func() {
				BeforeEach(func() {
					configRepo.SetAPIVersion("2.37.0")
				})

				It("requests the set_roles_by_username flag", func() {
					cmd.Requirements(factory, flagContext)
					Expect(flagRepo.FindByNameCallCount()).To(Equal(1))
					Expect(flagRepo.FindByNameArgsForCall(0)).To(Equal("set_roles_by_username"))
				})

				Context("when the set_roles_by_username flag exists and is enabled", func() {
					BeforeEach(func() {
						flagRepo.FindByNameReturns(models.FeatureFlag{Enabled: true}, nil)
					})

					It("returns a UserRequirement", func() {
						actualRequirements, err := cmd.Requirements(factory, flagContext)
						Expect(err).NotTo(HaveOccurred())
						Expect(factory.NewUserRequirementCallCount()).To(Equal(1))
						actualUsername, actualWantGUID := factory.NewUserRequirementArgsForCall(0)
						Expect(actualUsername).To(Equal("the-user-name"))
						Expect(actualWantGUID).To(BeFalse())

						Expect(actualRequirements).To(ContainElement(userRequirement))
					})
				})

				Context("when the set_roles_by_username flag exists and is disabled", func() {
					BeforeEach(func() {
						flagRepo.FindByNameReturns(models.FeatureFlag{Enabled: false}, nil)
					})

					It("returns a UserRequirement", func() {
						actualRequirements, err := cmd.Requirements(factory, flagContext)
						Expect(err).NotTo(HaveOccurred())
						Expect(factory.NewUserRequirementCallCount()).To(Equal(1))
						actualUsername, actualWantGUID := factory.NewUserRequirementArgsForCall(0)
						Expect(actualUsername).To(Equal("the-user-name"))
						Expect(actualWantGUID).To(BeTrue())

						Expect(actualRequirements).To(ContainElement(userRequirement))
					})
				})

				Context("when the set_roles_by_username flag cannot be retrieved", func() {
					BeforeEach(func() {
						flagRepo.FindByNameReturns(models.FeatureFlag{}, errors.New("some error"))
					})

					It("returns a UserRequirement", func() {
						actualRequirements, err := cmd.Requirements(factory, flagContext)
						Expect(err).NotTo(HaveOccurred())
						Expect(factory.NewUserRequirementCallCount()).To(Equal(1))
						actualUsername, actualWantGUID := factory.NewUserRequirementArgsForCall(0)
						Expect(actualUsername).To(Equal("the-user-name"))
						Expect(actualWantGUID).To(BeTrue())

						Expect(actualRequirements).To(ContainElement(userRequirement))
					})
				})
			})

			Context("when the config version is <2.37.0", func() {
				BeforeEach(func() {
					configRepo.SetAPIVersion("2.36.0")
				})

				It("returns a UserRequirement", func() {
					actualRequirements, err := cmd.Requirements(factory, flagContext)
					Expect(err).NotTo(HaveOccurred())
					Expect(factory.NewUserRequirementCallCount()).To(Equal(1))
					actualUsername, actualWantGUID := factory.NewUserRequirementArgsForCall(0)
					Expect(actualUsername).To(Equal("the-user-name"))
					Expect(actualWantGUID).To(BeTrue())

					Expect(actualRequirements).To(ContainElement(userRequirement))
				})
			})
		})
	})

	Describe("Execute", func() {
		var (
			org models.Organization
			err error
		)

		BeforeEach(func() {
			flagContext.Parse("the-user-name", "the-org-name", "the-space-name", "SpaceManager")
			cmd.Requirements(factory, flagContext)

			org = models.Organization{}
			org.GUID = "the-org-guid"
			org.Name = "the-org-name"
			organizationRequirement.GetOrganizationReturns(org)
		})

		JustBeforeEach(func() {
			err = cmd.Execute(flagContext)
		})

		Context("when the space is not found", func() {
			BeforeEach(func() {
				spaceRepo.FindByNameInOrgReturns(models.Space{}, errors.New("space-repo-error"))
			})

			It("doesn't call CC", func() {
				Expect(userRepo.SetSpaceRoleByGUIDCallCount()).To(BeZero())
				Expect(userRepo.SetSpaceRoleByUsernameCallCount()).To(BeZero())
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("space-repo-error"))
			})
		})

		Context("when the space is found", func() {
			BeforeEach(func() {
				space := models.Space{}
				space.GUID = "the-space-guid"
				space.Name = "the-space-name"
				spaceRepo.FindByNameInOrgReturns(space, nil)
			})

			Context("when the UserRequirement returns a user with a GUID", func() {
				BeforeEach(func() {
					userFields := models.UserFields{GUID: "the-user-guid", Username: "the-user-name"}
					userRequirement.GetUserReturns(userFields)
				})

				It("tells the user it is assigning the role", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Assigning role", "SpaceManager", "the-user-name", "the-org", "the-user-name"},
						[]string{"OK"},
					))
				})

				It("sets the role using the GUID", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(userRepo.SetSpaceRoleByGUIDCallCount()).To(Equal(1))
					actualUserGUID, actualSpaceGUID, actualOrgGUID, actualRole := userRepo.SetSpaceRoleByGUIDArgsForCall(0)
					Expect(actualUserGUID).To(Equal("the-user-guid"))
					Expect(actualSpaceGUID).To(Equal("the-space-guid"))
					Expect(actualOrgGUID).To(Equal("the-org-guid"))
					Expect(actualRole).To(Equal(models.RoleSpaceManager))
				})

				Context("when the call to CC fails", func() {
					BeforeEach(func() {
						userRepo.SetSpaceRoleByGUIDReturns(errors.New("user-repo-error"))
					})

					It("returns an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("user-repo-error"))
					})
				})
			})

			Context("when the UserRequirement returns a user without a GUID", func() {
				BeforeEach(func() {
					userRequirement.GetUserReturns(models.UserFields{Username: "the-user-name"})
				})

				It("sets the role using the given username", func() {
					Expect(err).NotTo(HaveOccurred())
					username, spaceGUID, orgGUID, role := userRepo.SetSpaceRoleByUsernameArgsForCall(0)
					Expect(username).To(Equal("the-user-name"))
					Expect(spaceGUID).To(Equal("the-space-guid"))
					Expect(orgGUID).To(Equal("the-org-guid"))
					Expect(role).To(Equal(models.RoleSpaceManager))
				})

				It("tells the user it assigned the role", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Assigning role", "SpaceManager", "the-user-name", "the-org", "the-user-name"},
						[]string{"OK"},
					))
				})

				Context("when the call to CC fails", func() {
					BeforeEach(func() {
						userRepo.SetSpaceRoleByUsernameReturns(errors.New("user-repo-error"))
					})

					It("returns an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("user-repo-error"))
					})
				})
			})
		})
	})
})
