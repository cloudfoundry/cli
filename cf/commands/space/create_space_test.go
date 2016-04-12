package space_test

import (
	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/api/feature_flags/featureflagsfakes"
	"github.com/cloudfoundry/cli/cf/api/organizations/organizationsfakes"
	"github.com/cloudfoundry/cli/cf/api/spacequotas/spacequotasfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
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
		configRepo          coreconfig.Repository
		spaceRepo           *apifakes.FakeSpaceRepository
		orgRepo             *organizationsfakes.FakeOrganizationRepository
		userRepo            *apifakes.FakeUserRepository
		spaceRoleSetter     user.SpaceRoleSetter
		flagRepo            *featureflagsfakes.FakeFeatureFlagRepository
		spaceQuotaRepo      *spacequotasfakes.FakeSpaceQuotaRepository
		OriginalCommand     commandregistry.Command
		deps                commandregistry.Dependency
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
		commandregistry.Register(spaceRoleSetter)

		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("create-space").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("create-space", args, requirementsFactory, updateCommandDependency, false)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()

		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		userRepo = new(apifakes.FakeUserRepository)
		spaceRoleSetter = commandregistry.Commands.FindCommand("set-space-role").(user.SpaceRoleSetter)
		spaceQuotaRepo = new(spacequotasfakes.FakeSpaceQuotaRepository)
		flagRepo = new(featureflagsfakes.FakeFeatureFlagRepository)

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		configOrg = models.OrganizationFields{
			Name: "my-org",
			Guid: "my-org-guid",
		}

		//save original command and restore later
		OriginalCommand = commandregistry.Commands.FindCommand("set-space-role")

		spaceRepo = new(apifakes.FakeSpaceRepository)
		space := maker.NewSpace(maker.Overrides{"name": "my-space", "guid": "my-space-guid", "organization": configOrg})
		spaceRepo.CreateReturns(space, nil)
	})

	AfterEach(func() {
		commandregistry.Register(OriginalCommand)
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
		spaceRepo.CreateReturns(models.Space{}, errors.NewHttpError(400, errors.SpaceNameTaken, "Space already exists"))
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
			spaceQuotaRepo.FindByNameAndOrgGuidReturns(spaceQuota, nil)
			runCommand("-q", "my-space-quota", "my-space")

			spaceQuotaName, orgGuid := spaceQuotaRepo.FindByNameAndOrgGuidArgsForCall(0)

			Expect(spaceQuotaName).To(Equal(spaceQuota.Name))
			Expect(orgGuid).To(Equal(configOrg.Guid))

			_, _, actualSpaceQuotaGUID := spaceRepo.CreateArgsForCall(0)
			Expect(actualSpaceQuotaGUID).To(Equal(spaceQuota.Guid))
		})

		Context("when the space-quota provided does not exist", func() {
			It("fails", func() {
				spaceQuotaRepo.FindByNameAndOrgGuidReturns(models.SpaceQuota{}, errors.New("Error"))
				runCommand("-q", "my-space-quota", "my-space")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Error"},
				))
			})
		})
	})

	Context("when the -o and -q flags are provided", func() {
		BeforeEach(func() {
			org := models.Organization{
				OrganizationFields: models.OrganizationFields{
					Name: "other-org",
					Guid: "other-org-guid",
				}}
			orgRepo.FindByNameReturns(org, nil)

			spaceQuota := models.SpaceQuota{
				Name: "my-space-quota",
				Guid: "my-space-quota-guid",
			}
			spaceQuotaRepo.FindByNameAndOrgGuidReturns(spaceQuota, nil)
		})

		It("assigns the space-quota from the specified org to the space", func() {
			runCommand("my-space", "-o", "other-org", "-q", "my-space-quota")

			Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("other-org"))
			spaceName, orgGuid := spaceQuotaRepo.FindByNameAndOrgGuidArgsForCall(0)
			Expect(spaceName).To(Equal("my-space-quota"))
			Expect(orgGuid).To(Equal("other-org-guid"))

			actualSpaceName, actualOrgGuid, actualSpaceQuotaGUID := spaceRepo.CreateArgsForCall(0)
			Expect(actualSpaceName).To(Equal("my-space"))
			Expect(actualOrgGuid).To(Equal("other-org-guid"))
			Expect(actualSpaceQuotaGUID).To(Equal("my-space-quota-guid"))
		})
	})
})
