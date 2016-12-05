package commands_test

import (
	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/api/spaces/spacesfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commands"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("target command", func() {
	var (
		orgRepo             *organizationsfakes.FakeOrganizationRepository
		spaceRepo           *spacesfakes.FakeSpaceRepository
		requirementsFactory *requirementsfakes.FakeFactory
		config              coreconfig.Repository
		ui                  *testterm.FakeUI
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("target").SetDependency(deps, pluginCall))
	}

	listSpacesStub := func(spaces []models.Space) func(func(models.Space) bool) error {
		return func(cb func(models.Space) bool) error {
			var keepGoing bool
			for _, s := range spaces {
				keepGoing = cb(s)
				if !keepGoing {
					break
				}
			}
			return nil
		}
	}

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		spaceRepo = new(spacesfakes.FakeSpaceRepository)
		config = testconfig.NewRepository()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewAPIEndpointRequirementReturns(requirements.Passing{})
	})

	var callTarget = func(args []string) bool {
		return testcmd.RunCLICommand("target", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Context("when there are too many arguments given", func() {
		var cmd commandregistry.Command
		var flagContext flags.FlagContext

		BeforeEach(func() {
			cmd = new(commands.Target)
			cmd.SetDependency(deps, false)
			flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		})

		It("fails with usage", func() {
			flagContext.Parse("blahblah")

			reqs, err := cmd.Requirements(requirementsFactory, flagContext)
			Expect(err).NotTo(HaveOccurred())

			err = testcmd.RunRequirements(reqs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Incorrect Usage"))
			Expect(err.Error()).To(ContainSubstring("No argument required"))
		})
	})

	Describe("when there is no api endpoint set", func() {
		BeforeEach(func() {
			requirementsFactory.NewAPIEndpointRequirementReturns(requirements.Failing{Message: "no api set"})
		})

		It("fails requirements", func() {
			Expect(callTarget([]string{})).To(BeFalse())
		})
	})

	Describe("when the user is not logged in", func() {
		BeforeEach(func() {
			config.SetAccessToken("")
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
		})

		It("prints the target info when no org or space is specified", func() {
			Expect(callTarget([]string{})).To(BeFalse())
			Expect(ui.ShowConfigurationCalled).To(BeTrue())
		})

		It("fails requirements when targeting a space or org", func() {
			Expect(callTarget([]string{"-o", "some-crazy-org-im-not-in"})).To(BeFalse())

			Expect(callTarget([]string{"-s", "i-love-space"})).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		})

		var expectOrgToBeCleared = func() {
			Expect(config.OrganizationFields()).To(Equal(models.OrganizationFields{}))
		}

		var expectSpaceToBeCleared = func() {
			Expect(config.SpaceFields()).To(Equal(models.SpaceFields{}))
		}

		Context("there are no errors", func() {
			BeforeEach(func() {
				org := models.Organization{}
				org.Name = "my-organization"
				org.GUID = "my-organization-guid"

				orgRepo.ListOrgsReturns([]models.Organization{org}, nil)
				orgRepo.FindByNameReturns(org, nil)
				config.SetOrganizationFields(models.OrganizationFields{Name: org.Name, GUID: org.GUID})
			})

			It("it updates the organization in the config", func() {
				callTarget([]string{"-o", "my-organization"})

				Expect(orgRepo.FindByNameCallCount()).To(Equal(1))
				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-organization"))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())

				Expect(config.OrganizationFields().GUID).To(Equal("my-organization-guid"))
			})

			It("updates the space in the config", func() {
				space := models.Space{}
				space.Name = "my-space"
				space.GUID = "my-space-guid"

				spaceRepo.FindByNameReturns(space, nil)

				callTarget([]string{"-s", "my-space"})

				Expect(spaceRepo.FindByNameCallCount()).To(Equal(1))
				Expect(spaceRepo.FindByNameArgsForCall(0)).To(Equal("my-space"))
				Expect(config.SpaceFields().GUID).To(Equal("my-space-guid"))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("updates both the organization and the space in the config", func() {
				space := models.Space{}
				space.Name = "my-space"
				space.GUID = "my-space-guid"
				spaceRepo.FindByNameReturns(space, nil)

				callTarget([]string{"-o", "my-organization", "-s", "my-space"})

				Expect(orgRepo.FindByNameCallCount()).To(Equal(1))
				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-organization"))
				Expect(config.OrganizationFields().GUID).To(Equal("my-organization-guid"))

				Expect(spaceRepo.FindByNameCallCount()).To(Equal(1))
				Expect(spaceRepo.FindByNameArgsForCall(0)).To(Equal("my-space"))
				Expect(config.SpaceFields().GUID).To(Equal("my-space-guid"))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("only updates the organization in the config when the space can't be found", func() {
				config.SetSpaceFields(models.SpaceFields{})

				spaceRepo.FindByNameReturns(models.Space{}, errors.New("Error finding space by name."))

				callTarget([]string{"-o", "my-organization", "-s", "my-space"})

				Expect(orgRepo.FindByNameCallCount()).To(Equal(1))
				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-organization"))
				Expect(config.OrganizationFields().GUID).To(Equal("my-organization-guid"))

				Expect(spaceRepo.FindByNameCallCount()).To(Equal(1))
				Expect(spaceRepo.FindByNameArgsForCall(0)).To(Equal("my-space"))
				Expect(config.SpaceFields().GUID).To(Equal(""))

				Expect(ui.ShowConfigurationCalled).To(BeFalse())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Unable to access space", "my-space"},
				))
			})

			Describe("when there is only a single space", func() {
				It("target space automatically ", func() {
					space := models.Space{}
					space.Name = "my-space"
					space.GUID = "my-space-guid"
					spaceRepo.FindByNameReturns(space, nil)
					spaceRepo.ListSpacesStub = listSpacesStub([]models.Space{space})

					callTarget([]string{"-o", "my-organization"})

					Expect(config.OrganizationFields().GUID).To(Equal("my-organization-guid"))
					Expect(config.SpaceFields().GUID).To(Equal("my-space-guid"))

					Expect(ui.ShowConfigurationCalled).To(BeTrue())
				})

			})

			It("not target space automatically for orgs having multiple spaces", func() {
				space1 := models.Space{}
				space1.Name = "my-space"
				space1.GUID = "my-space-guid"
				space2 := models.Space{}
				space2.Name = "my-space"
				space2.GUID = "my-space-guid"
				spaceRepo.ListSpacesStub = listSpacesStub([]models.Space{space1, space2})

				callTarget([]string{"-o", "my-organization"})

				Expect(config.OrganizationFields().GUID).To(Equal("my-organization-guid"))
				Expect(config.SpaceFields().GUID).To(Equal(""))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("displays an update notification", func() {
				callTarget([]string{"-o", "my-organization"})
				Expect(ui.NotifyUpdateIfNeededCallCount).To(Equal(1))
			})
		})

		Context("there are errors", func() {
			It("fails when the user does not have access to the specified organization", func() {
				orgRepo.FindByNameReturns(models.Organization{}, errors.New("Invalid access"))

				callTarget([]string{"-o", "my-organization"})
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
				expectOrgToBeCleared()
				expectSpaceToBeCleared()
			})

			It("fails when the organization is not found", func() {
				orgRepo.FindByNameReturns(models.Organization{}, errors.New("my-organization not found"))

				callTarget([]string{"-o", "my-organization"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"my-organization", "not found"},
				))

				expectOrgToBeCleared()
				expectSpaceToBeCleared()
			})

			It("fails to target a space if no organization is targeted", func() {
				config.SetOrganizationFields(models.OrganizationFields{})

				callTarget([]string{"-s", "my-space"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"An org must be targeted before targeting a space"},
				))

				expectSpaceToBeCleared()
			})

			Context("when an org is targetted", func() {
				var org models.Organization
				BeforeEach(func() {
					org = models.Organization{}
					org.Name = "my-organization"
					org.GUID = "my-organization-guid"

					config.SetOrganizationFields(models.OrganizationFields{Name: org.Name, GUID: org.GUID})
				})

				It("fails when the user doesn't have access to the space", func() {
					spaceRepo.FindByNameReturns(models.Space{}, errors.New("Error finding space by name."))

					callTarget([]string{"-s", "my-space"})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Unable to access space", "my-space"},
					))

					Expect(config.SpaceFields().GUID).To(Equal(""))
					Expect(ui.ShowConfigurationCalled).To(BeFalse())

					Expect(config.OrganizationFields().GUID).NotTo(BeEmpty())
					expectSpaceToBeCleared()
				})

				It("fails when the space is not found", func() {
					spaceRepo.FindByNameReturns(models.Space{}, errors.NewModelNotFoundError("Space", "my-space"))

					callTarget([]string{"-s", "my-space"})

					expectSpaceToBeCleared()
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"my-space", "not found"},
					))
				})

				It("fails to target the space automatically if is not found", func() {
					orgRepo.ListOrgsReturns([]models.Organization{org}, nil)
					orgRepo.FindByNameReturns(org, nil)

					spaceRepo.FindByNameReturns(models.Space{}, errors.NewModelNotFoundError("Space", "my-space"))

					callTarget([]string{"-o", "my-organization"})

					Expect(config.OrganizationFields().GUID).To(Equal("my-organization-guid"))
					Expect(config.SpaceFields().GUID).To(Equal(""))

					Expect(ui.ShowConfigurationCalled).To(BeTrue())
				})
			})
		})
	})
})
