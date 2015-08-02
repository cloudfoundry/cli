package space_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fake_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/api/space_quotas/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/testhelpers/maker"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-space command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		configSpace         models.SpaceFields
		configOrg           models.OrganizationFields
		configRepo          core_config.Repository
		spaceRepo           *testapi.FakeSpaceRepository
		orgRepo             *fake_org.FakeOrganizationRepository
		userRepo            *testapi.FakeUserRepository
		spaceRoleSetter     user.SpaceRoleSetter
		spaceQuotaRepo      *fakes.FakeSpaceQuotaRepository
		OriginalCommand     command_registry.Command
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		deps.RepoLocator = deps.RepoLocator.SetSpaceQuotaRepository(spaceQuotaRepo)
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.RepoLocator = deps.RepoLocator.SetUserRepository(userRepo)
		deps.Config = configRepo

		//inject fake 'command dependency' into registry
		command_registry.Register(spaceRoleSetter)

		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("create-space").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("create-space", args, requirementsFactory, updateCommandDependency, false)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()

		orgRepo = &fake_org.FakeOrganizationRepository{}
		userRepo = &testapi.FakeUserRepository{}
		spaceRoleSetter = command_registry.Commands.FindCommand("set-space-role").(user.SpaceRoleSetter)
		spaceQuotaRepo = &fakes.FakeSpaceQuotaRepository{}

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		configOrg = models.OrganizationFields{
			Name: "my-org",
			Guid: "my-org-guid",
		}

		configSpace = models.SpaceFields{
			Name: "config-space",
			Guid: "config-space-guid",
		}

		//save original command and restore later
		OriginalCommand = command_registry.Commands.FindCommand("set-space-role")

		spaceRepo = &testapi.FakeSpaceRepository{
			CreateSpaceSpace: maker.NewSpace(maker.Overrides{"name": "my-space", "guid": "my-space-guid", "organization": configOrg}),
		}
		Expect(spaceRepo.CreateSpaceSpace.Name).To(Equal("my-space"))
	})

	AfterEach(func() {
		command_registry.Register(OriginalCommand)
	})

	Describe("Requirements", func() {
		It("fails with usage when not provided exactly one argument", func() {
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				requirementsFactory.LoginSuccess = false
			})

			It("fails requirements", func() {
				Expect(runCommand("some-space")).To(BeFalse())
			})
		})

		Context("when a org is not targeted", func() {
			BeforeEach(func() {
				requirementsFactory.TargetedOrgSuccess = false
			})

			It("fails requirements", func() {
				Expect(runCommand("what-is-space?")).To(BeFalse())
			})
		})
	})

	It("creates a space", func() {
		runCommand("my-space")
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating space", "my-space", "my-org", "my-user"},
			[]string{"OK"},
			[]string{"Assigning", models.SpaceRoleToUserInput[models.SPACE_MANAGER], "my-user", "my-space"},
			[]string{"Assigning", models.SpaceRoleToUserInput[models.SPACE_DEVELOPER], "my-user", "my-space"},
			[]string{"TIP"},
		))

		Expect(spaceRepo.CreateSpaceName).To(Equal("my-space"))
		Expect(spaceRepo.CreateSpaceOrgGuid).To(Equal("my-org-guid"))
		Expect(userRepo.SetSpaceRoleUserGuid).To(Equal("my-user-guid"))
		Expect(userRepo.SetSpaceRoleSpaceGuid).To(Equal("my-space-guid"))
		Expect(userRepo.SetSpaceRoleRole).To(Equal(models.SPACE_DEVELOPER))
	})

	It("warns the user when a space with that name already exists", func() {
		spaceRepo.CreateSpaceExists = true
		runCommand("my-space")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating space", "my-space"},
			[]string{"OK"},
		))
		Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"my-space", "already exists"}))
		Expect(ui.Outputs).ToNot(ContainSubstrings(
			[]string{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_MANAGER]},
		))

		Expect(spaceRepo.CreateSpaceName).To(Equal(""))
		Expect(spaceRepo.CreateSpaceOrgGuid).To(Equal(""))
		Expect(userRepo.SetSpaceRoleUserGuid).To(Equal(""))
		Expect(userRepo.SetSpaceRoleSpaceGuid).To(Equal(""))
	})

	Context("when the -o flag is provided", func() {
		It("creates a space within that org", func() {
			org := models.Organization{
				OrganizationFields: models.OrganizationFields{
					Name: "other-org",
					Guid: "org-guid-1",
				}}
			orgRepo.FindByNameReturns(org, nil)

			runCommand("-o", "other-org", "my-space")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating space", "my-space", "other-org", "my-user"},
				[]string{"OK"},
				[]string{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_MANAGER]},
				[]string{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_DEVELOPER]},
				[]string{"TIP"},
			))

			Expect(spaceRepo.CreateSpaceName).To(Equal("my-space"))
			Expect(spaceRepo.CreateSpaceOrgGuid).To(Equal(org.Guid))
			Expect(userRepo.SetSpaceRoleUserGuid).To(Equal("my-user-guid"))
			Expect(userRepo.SetSpaceRoleSpaceGuid).To(Equal("my-space-guid"))
			Expect(userRepo.SetSpaceRoleRole).To(Equal(models.SPACE_DEVELOPER))
		})

		It("fails when the org provided does not exist", func() {
			orgRepo.FindByNameReturns(models.Organization{}, errors.New("cool-organization does not exist"))
			runCommand("-o", "cool-organization", "my-space")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"cool-organization", "does not exist"},
			))

			Expect(spaceRepo.CreateSpaceName).To(Equal(""))
		})

		It("fails when finding the org returns an error", func() {
			orgRepo.FindByNameReturns(models.Organization{}, errors.New("cool-organization does not exist"))
			runCommand("-o", "cool-organization", "my-space")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"Error"},
			))

			Expect(spaceRepo.CreateSpaceName).To(Equal(""))
		})
	})

	Context("when the -q flag is provided", func() {
		It("assigns the space-quota specified to the space", func() {
			spaceQuota := models.SpaceQuota{
				Name: "my-space-quota",
				Guid: "my-space-quota-guid",
			}
			spaceQuotaRepo.FindByNameReturns(spaceQuota, nil)
			runCommand("-q", "my-space-quota", "my-space")

			Expect(spaceQuotaRepo.FindByNameArgsForCall(0)).To(Equal(spaceQuota.Name))
			Expect(spaceRepo.CreateSpaceSpaceQuotaGuid).To(Equal(spaceQuota.Guid))

		})

		Context("when the space-quota provided does not exist", func() {
			It("fails", func() {
				spaceQuotaRepo.FindByNameReturns(models.SpaceQuota{}, errors.New("Error"))
				runCommand("-q", "my-space-quota", "my-space")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Error"},
				))
			})
		})
	})
})
