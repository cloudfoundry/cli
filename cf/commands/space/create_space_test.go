package space_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakeflag "github.com/cloudfoundry/cli/cf/api/feature_flags/fakes"
	fake_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/api/space_quotas/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
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
		configOrg           models.OrganizationFields
		configRepo          core_config.Repository
		spaceRepo           *testapi.FakeSpaceRepository
		orgRepo             *fake_org.FakeOrganizationRepository
		userRepo            *testapi.FakeUserRepository
		spaceRoleSetter     user.SpaceRoleSetter
		flagRepo            *fakeflag.FakeFeatureFlagRepository
		spaceQuotaRepo      *fakes.FakeSpaceQuotaRepository
		OriginalCommand     command_registry.Command
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		deps.RepoLocator = deps.RepoLocator.SetSpaceQuotaRepository(spaceQuotaRepo)
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.RepoLocator = deps.RepoLocator.SetFeatureFlagRepository(flagRepo)
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
		flagRepo = &fakeflag.FakeFeatureFlagRepository{}

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		configOrg = models.OrganizationFields{
			Name: "my-org",
			Guid: "my-org-guid",
		}

		//save original command and restore later
		OriginalCommand = command_registry.Commands.FindCommand("set-space-role")

		spaceRepo = &testapi.FakeSpaceRepository{}
		space := maker.NewSpace(maker.Overrides{"name": "my-space", "guid": "my-space-guid", "organization": configOrg})
		spaceRepo.CreateReturns(space, nil)
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

		name, orgGUID, _ := spaceRepo.CreateArgsForCall(0)
		Expect(name).To(Equal("my-space"))
		Expect(orgGUID).To(Equal("my-org-guid"))

		userGuid, spaceGuid, orgGuid, role := userRepo.SetSpaceRoleByGuidArgsForCall(0)
		Expect(userGuid).To(Equal("my-user-guid"))
		Expect(spaceGuid).To(Equal("my-space-guid"))
		Expect(orgGuid).To(Equal("my-org-guid"))
		Expect(role).To(Equal(models.SPACE_MANAGER))
	})

	It("warns the user when a space with that name already exists", func() {
		spaceRepo.CreateReturns(models.Space{}, errors.NewHttpError(400, errors.SPACE_EXISTS, "Space already exists"))
		runCommand("my-space")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating space", "my-space"},
			[]string{"OK"},
		))
		Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"my-space", "already exists"}))
		Expect(ui.Outputs).ToNot(ContainSubstrings(
			[]string{"Assigning", "my-user", "my-space", models.SpaceRoleToUserInput[models.SPACE_MANAGER]},
		))

		Expect(spaceRepo.CreateCallCount()).To(Equal(1))
		actualSpaceName, _, _ := spaceRepo.CreateArgsForCall(0)
		Expect(actualSpaceName).To(Equal("my-space"))
		Expect(userRepo.SetSpaceRoleByGuidCallCount()).To(BeZero())
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

			actualSpaceName, actualOrgGUID, _ := spaceRepo.CreateArgsForCall(0)
			Expect(actualSpaceName).To(Equal("my-space"))
			Expect(actualOrgGUID).To(Equal(org.Guid))

			userGuid, spaceGuid, orgGuid, role := userRepo.SetSpaceRoleByGuidArgsForCall(0)
			Expect(userGuid).To(Equal("my-user-guid"))
			Expect(spaceGuid).To(Equal("my-space-guid"))
			Expect(orgGuid).To(Equal("my-org-guid"))
			Expect(role).To(Equal(models.SPACE_MANAGER))
		})

		It("fails when the org provided does not exist", func() {
			orgRepo.FindByNameReturns(models.Organization{}, errors.New("cool-organization does not exist"))
			runCommand("-o", "cool-organization", "my-space")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"cool-organization", "does not exist"},
			))

			Expect(spaceRepo.CreateCallCount()).To(BeZero())
		})

		It("fails when finding the org returns an error", func() {
			orgRepo.FindByNameReturns(models.Organization{}, errors.New("cool-organization does not exist"))
			runCommand("-o", "cool-organization", "my-space")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"Error"},
			))

			Expect(spaceRepo.CreateCallCount()).To(BeZero())
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
			_, _, actualSpaceQuotaGUID := spaceRepo.CreateArgsForCall(0)
			Expect(actualSpaceQuotaGUID).To(Equal(spaceQuota.Guid))
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
