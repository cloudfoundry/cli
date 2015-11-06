package user_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/simonleung8/flags"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetSpaceRole", func() {
	var (
		ui         *testterm.FakeUI
		configRepo core_config.Repository
		userRepo   *testapi.FakeUserRepository
		spaceRepo  *testapi.FakeSpaceRepository

		cmd                 command_registry.Command
		requirementsFactory *testreq.FakeReqFactory
		flagContext         flags.FlagContext
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		userRepo = &testapi.FakeUserRepository{}
		spaceRepo = &testapi.FakeSpaceRepository{}

		deps := command_registry.Dependency{}
		repoLocator := deps.RepoLocator
		repoLocator = repoLocator.SetUserRepository(userRepo)
		repoLocator = repoLocator.SetSpaceRepository(spaceRepo)

		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = repoLocator

		cmd = &user.SetSpaceRole{}
		cmd.SetDependency(deps, false)

		requirementsFactory = &testreq.FakeReqFactory{}
		flagContext = flags.NewFlagContext(map[string]flags.FlagSet{})
	})

	Describe("Requirements", func() {
		Context("when not provided exactly four args", func() {
			BeforeEach(func() {
				flagContext.Parse("the-username", "the-org-name", "the-space-name")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(requirementsFactory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Incorrect Usage. Requires USERNAME, ORG, SPACE, ROLE as arguments"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})
		})

		Context("when provided three args", func() {
			BeforeEach(func() {
				flagContext.Parse("the-username", "the-org-name", "the-space-name", "SpaceManager")
			})

			It("returns a LoginRequirement", func() {
				requirementsFactory.LoginSuccess = false
				requirementsFactory.UserRequirementFails = false
				requirementsFactory.SpaceRequirementFails = false
				actualRequirements, err := cmd.Requirements(requirementsFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())

				var failures int
				for _, req := range actualRequirements {
					ok := req.Execute()
					if !ok {
						failures = failures + 1
					}
				}

				Expect(failures).To(Equal(1))
			})

			It("returns a UserRequirement", func() {
				requirementsFactory.LoginSuccess = true
				requirementsFactory.UserRequirementFails = true
				requirementsFactory.SpaceRequirementFails = false
				actualRequirements, err := cmd.Requirements(requirementsFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())

				var failures int
				for _, req := range actualRequirements {
					ok := req.Execute()
					if !ok {
						failures = failures + 1
					}
				}

				Expect(failures).To(Equal(1))
			})

			It("returns an OrgRequirement", func() {
				requirementsFactory.LoginSuccess = true
				requirementsFactory.UserRequirementFails = false
				requirementsFactory.OrganizationRequirementFails = true
				actualRequirements, err := cmd.Requirements(requirementsFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())

				var failures int
				for _, req := range actualRequirements {
					ok := req.Execute()
					if !ok {
						failures = failures + 1
					}
				}

				Expect(failures).To(Equal(1))
			})
		})
	})

	Context("Execute", func() {
		var org models.Organization

		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.UserRequirementFails = false
			requirementsFactory.OrganizationRequirementFails = false

			flagContext.Parse("the-username", "the-org-name", "the-space-name", "SpaceManager")
			_, err := cmd.Requirements(requirementsFactory, flagContext)
			Expect(err).NotTo(HaveOccurred())

			org = models.Organization{}
			org.Guid = "the-org-guid"
			org.Name = "the-org-name"
			requirementsFactory.Organization = org

		})

		Context("when the space is not found", func() {
			BeforeEach(func() {
				spaceRepo.FindByNameInOrgReturns(models.Space{}, errors.New("space-repo-error"))
			})

			It("doesn't call CC", func() {
				Expect(userRepo.SetSpaceRoleByGuidCallCount()).To(BeZero())
				Expect(userRepo.SetSpaceRoleByUsernameCallCount()).To(BeZero())
			})

			It("panics and prints a failure message", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(BeInDisplayOrder(
					[]string{"FAILED"},
					[]string{"space-repo-error"},
				))
			})
		})

		Context("when the space is found", func() {
			BeforeEach(func() {
				space := models.Space{}
				space.Guid = "the-space-guid"
				space.Name = "the-space-name"
				space.Organization = org.OrganizationFields
				spaceRepo.FindByNameInOrgReturns(space, nil)
			})

			Context("when the UserRequirement returns a user with a GUID", func() {
				BeforeEach(func() {
					userFields := models.UserFields{Guid: "the-user-guid", Username: "the-username"}
					requirementsFactory.UserFields = userFields
				})

				It("tells the user it is assigning the role", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Assigning role", "SpaceManager", "the-username", "the-org", "the-username"},
						[]string{"OK"},
					))
				})

				It("sets the role using the GUID", func() {
					cmd.Execute(flagContext)
					Expect(userRepo.SetSpaceRoleByGuidCallCount()).To(Equal(1))
					actualUserGUID, actualSpaceGUID, actualOrgGUID, actualRole := userRepo.SetSpaceRoleByGuidArgsForCall(0)
					Expect(actualUserGUID).To(Equal("the-user-guid"))
					Expect(actualSpaceGUID).To(Equal("the-space-guid"))
					Expect(actualOrgGUID).To(Equal("the-org-guid"))
					Expect(actualRole).To(Equal("SpaceManager"))
				})

				Context("when the call to CC fails", func() {
					BeforeEach(func() {
						userRepo.SetSpaceRoleByGuidReturns(errors.New("user-repo-error"))
					})

					It("panics and prints a failure message", func() {
						Expect(func() { cmd.Execute(flagContext) }).To(Panic())
						Expect(ui.Outputs).To(BeInDisplayOrder(
							[]string{"FAILED"},
							[]string{"user-repo-error"},
						))
					})
				})
			})

			Context("when the UserRequirement returns a user without a GUID", func() {
				BeforeEach(func() {
					requirementsFactory.UserFields = models.UserFields{Username: "the-username"}
				})

				It("sets the role using the given username", func() {
					cmd.Execute(flagContext)
					username, spaceGUID, orgGUID, role := userRepo.SetSpaceRoleByUsernameArgsForCall(0)
					Expect(username).To(Equal("the-username"))
					Expect(spaceGUID).To(Equal("the-space-guid"))
					Expect(orgGUID).To(Equal("the-org-guid"))
					Expect(role).To(Equal("SpaceManager"))
				})

				It("tells the user it assigned the role", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Assigning role", "SpaceManager", "the-username", "the-org", "the-username"},
						[]string{"OK"},
					))
				})

				Context("when the call to CC fails", func() {
					BeforeEach(func() {
						userRepo.SetSpaceRoleByUsernameReturns(errors.New("user-repo-error"))
					})

					It("panics and prints a failure message", func() {
						Expect(func() { cmd.Execute(flagContext) }).To(Panic())
						Expect(ui.Outputs).To(BeInDisplayOrder(
							[]string{"FAILED"},
							[]string{"user-repo-error"},
						))
					})
				})
			})
		})

	})
})
