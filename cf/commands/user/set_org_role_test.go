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

var _ = Describe("SetOrgRole", func() {
	var (
		ui         *testterm.FakeUI
		configRepo core_config.Repository
		userRepo   *testapi.FakeUserRepository

		cmd                 command_registry.Command
		deps                command_registry.Dependency
		requirementsFactory *testreq.FakeReqFactory
		flagContext         flags.FlagContext
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		userRepo = &testapi.FakeUserRepository{}
		repoLocator := deps.RepoLocator.SetUserRepository(userRepo)

		deps = command_registry.Dependency{
			Ui:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &user.SetOrgRole{}
		cmd.SetDependency(deps, false)

		requirementsFactory = &testreq.FakeReqFactory{}
		flagContext = flags.NewFlagContext(map[string]flags.FlagSet{})
	})

	Describe("Requirements", func() {
		Context("when not provided exactly three args", func() {
			BeforeEach(func() {
				flagContext.Parse("the-username", "the-org-name")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(requirementsFactory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Incorrect Usage. Requires USERNAME, ORG, ROLE as arguments"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})
		})

		Context("when provided three args", func() {
			BeforeEach(func() {
				flagContext.Parse("the-username", "the-org-name", "OrgManager")
			})

			It("returns a LoginRequirement", func() {
				requirementsFactory.LoginSuccess = false
				requirementsFactory.UserRequirementFails = false
				requirementsFactory.OrganizationRequirementFails = false
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
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.UserRequirementFails = false
			requirementsFactory.OrganizationRequirementFails = false

			flagContext.Parse("the-username", "the-org-name", "OrgManager")
			_, err := cmd.Requirements(requirementsFactory, flagContext)
			Expect(err).NotTo(HaveOccurred())

			org := models.Organization{}
			org.Guid = "the-org-guid"
			org.Name = "the-org-name"
			requirementsFactory.Organization = org
		})

		Context("when the UserRequirement returns a user with a GUID", func() {
			BeforeEach(func() {
				userFields := models.UserFields{Guid: "the-user-guid", Username: "the-username"}
				requirementsFactory.UserFields = userFields
			})

			It("tells the user it is assigning the role", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Assigning role", "OrgManager", "the-username", "the-org", "the-username"},
					[]string{"OK"},
				))
			})

			It("sets the role using the GUID", func() {
				cmd.Execute(flagContext)
				Expect(userRepo.SetOrgRoleCallCount()).To(Equal(1))
				actualUserGUID, actualOrgGUID, actualRole := userRepo.SetOrgRoleArgsForCall(0)
				Expect(actualUserGUID).To(Equal("the-user-guid"))
				Expect(actualOrgGUID).To(Equal("the-org-guid"))
				Expect(actualRole).To(Equal("OrgManager"))
			})

			Context("when the call to CC fails", func() {
				BeforeEach(func() {
					userRepo.SetOrgRoleReturns(errors.New("user-repo-error"))
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
				username, orgGuid, role := userRepo.SetOrgRoleByUsernameArgsForCall(0)
				Expect(username).To(Equal("the-username"))
				Expect(orgGuid).To(Equal("the-org-guid"))
				Expect(role).To(Equal("OrgManager"))
			})

			It("tells the user it assigned the role", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Assigning role", "OrgManager", "the-username", "the-org", "the-username"},
					[]string{"OK"},
				))
			})

			Context("when the call to CC fails", func() {
				BeforeEach(func() {
					userRepo.SetOrgRoleByUsernameReturns(errors.New("user-repo-error"))
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
