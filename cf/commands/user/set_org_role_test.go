package user_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"

	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/api/featureflags/featureflagsfakes"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetOrgRole", func() {
	var (
		ui         *testterm.FakeUI
		configRepo coreconfig.Repository
		userRepo   *apifakes.FakeUserRepository
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
		flagRepo = new(featureflagsfakes.FakeFeatureFlagRepository)
		repoLocator = repoLocator.SetFeatureFlagRepository(flagRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &user.SetOrgRole{}
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
		Context("when not provided exactly three args", func() {
			BeforeEach(func() {
				flagContext.Parse("the-user-name", "the-org-name")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(factory, flagContext) }).To(Panic())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Incorrect Usage. Requires USERNAME, ORG, ROLE as arguments"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})
		})

		Context("when provided three args", func() {
			BeforeEach(func() {
				flagContext.Parse("the-user-name", "the-org-name", "OrgManager")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))

				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns an OrgRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
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
						actualRequirements := cmd.Requirements(factory, flagContext)
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
						actualRequirements := cmd.Requirements(factory, flagContext)
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
						actualRequirements := cmd.Requirements(factory, flagContext)
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
					actualRequirements := cmd.Requirements(factory, flagContext)
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
		var err error

		BeforeEach(func() {
			flagContext.Parse("the-user-name", "the-org-name", "OrgManager")
			cmd.Requirements(factory, flagContext)

			org := models.Organization{}
			org.GUID = "the-org-guid"
			org.Name = "the-org-name"
			organizationRequirement.GetOrganizationReturns(org)
		})

		JustBeforeEach(func() {
			err = cmd.Execute(flagContext)
		})

		Context("when the UserRequirement returns a user with a GUID", func() {
			BeforeEach(func() {
				userFields := models.UserFields{GUID: "the-user-guid", Username: "the-user-name"}
				userRequirement.GetUserReturns(userFields)
			})

			It("tells the user it is assigning the role", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Assigning role", "OrgManager", "the-user-name", "the-org", "the-user-name"},
					[]string{"OK"},
				))
			})

			It("sets the role using the GUID", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(userRepo.SetOrgRoleByGUIDCallCount()).To(Equal(1))
				actualUserGUID, actualOrgGUID, actualRole := userRepo.SetOrgRoleByGUIDArgsForCall(0)
				Expect(actualUserGUID).To(Equal("the-user-guid"))
				Expect(actualOrgGUID).To(Equal("the-org-guid"))
				Expect(actualRole).To(Equal(models.RoleOrgManager))
			})

			Context("when the call to CC fails", func() {
				BeforeEach(func() {
					userRepo.SetOrgRoleByGUIDReturns(errors.New("user-repo-error"))
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
				username, orgGUID, role := userRepo.SetOrgRoleByUsernameArgsForCall(0)
				Expect(username).To(Equal("the-user-name"))
				Expect(orgGUID).To(Equal("the-org-guid"))
				Expect(role).To(Equal(models.RoleOrgManager))
			})

			It("tells the user it assigned the role", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Assigning role", "OrgManager", "the-user-name", "the-org", "the-user-name"},
					[]string{"OK"},
				))
			})

			Context("when the call to CC fails", func() {
				BeforeEach(func() {
					userRepo.SetOrgRoleByUsernameReturns(errors.New("user-repo-error"))
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("user-repo-error"))
				})
			})
		})
	})
})
