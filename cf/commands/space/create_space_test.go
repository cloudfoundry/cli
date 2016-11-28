package space_test

import (
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/api/featureflags/featureflagsfakes"
	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/api/spacequotas/spacequotasfakes"
	"code.cloudfoundry.org/cli/cf/api/spaces/spacesfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/user"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-space command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		configOrg           models.OrganizationFields
		configRepo          coreconfig.Repository
		spaceRepo           *spacesfakes.FakeSpaceRepository
		orgRepo             *organizationsfakes.FakeOrganizationRepository
		userRepo            *apifakes.FakeUserRepository
		spaceRoleSetter     user.SpaceRoleSetter
		flagRepo            *featureflagsfakes.FakeFeatureFlagRepository
		spaceQuotaRepo      *spacequotasfakes.FakeSpaceQuotaRepository
		OriginalCommand     commandregistry.Command
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
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
		return testcmd.RunCLICommand("create-space", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()

		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		userRepo = new(apifakes.FakeUserRepository)
		spaceRoleSetter = commandregistry.Commands.FindCommand("set-space-role").(user.SpaceRoleSetter)
		spaceQuotaRepo = new(spacequotasfakes.FakeSpaceQuotaRepository)
		flagRepo = new(featureflagsfakes.FakeFeatureFlagRepository)

		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
		configOrg = models.OrganizationFields{
			Name: "my-org",
			GUID: "my-org-guid",
		}

		//save original command and restore later
		OriginalCommand = commandregistry.Commands.FindCommand("set-space-role")

		spaceRepo = new(spacesfakes.FakeSpaceRepository)
		space := models.Space{SpaceFields: models.SpaceFields{
			Name: "my-space",
			GUID: "my-space-guid",
		}}
		spaceRepo.CreateReturns(space, nil)
	})

	AfterEach(func() {
		commandregistry.Register(OriginalCommand)
	})

	Describe("Requirements", func() {
		It("fails with usage when not provided exactly one argument", func() {
			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			})

			It("fails requirements", func() {
				Expect(runCommand("some-space")).To(BeFalse())
			})
		})

		Context("when a org is not targeted", func() {
			BeforeEach(func() {
				targetedOrgReq := new(requirementsfakes.FakeTargetedOrgRequirement)
				targetedOrgReq.ExecuteReturns(errors.New("no org targeted"))
				requirementsFactory.NewTargetedOrgRequirementReturns(targetedOrgReq)
			})

			It("fails requirements", func() {
				Expect(runCommand("what-is-space?")).To(BeFalse())
			})
		})
	})

	It("creates a space", func() {
		runCommand("my-space")
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Creating space", "my-space", "my-org", "my-user"},
			[]string{"OK"},
			[]string{"Assigning", "SpaceManager", "my-user", "my-space"},
			[]string{"Assigning", "SpaceDeveloper", "my-user", "my-space"},
			[]string{"TIP"},
		))

		name, orgGUID, _ := spaceRepo.CreateArgsForCall(0)
		Expect(name).To(Equal("my-space"))
		Expect(orgGUID).To(Equal("my-org-guid"))

		userGUID, spaceGUID, orgGUID, role := userRepo.SetSpaceRoleByGUIDArgsForCall(0)
		Expect(userGUID).To(Equal("my-user-guid"))
		Expect(spaceGUID).To(Equal("my-space-guid"))
		Expect(orgGUID).To(Equal("my-org-guid"))
		Expect(role).To(Equal(models.RoleSpaceManager))
	})

	It("warns the user when a space with that name already exists", func() {
		spaceRepo.CreateReturns(models.Space{}, errors.NewHTTPError(400, errors.SpaceNameTaken, "Space already exists"))
		runCommand("my-space")

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Creating space", "my-space"},
			[]string{"OK"},
		))
		Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"my-space", "already exists"}))
		Expect(ui.Outputs()).ToNot(ContainSubstrings(
			[]string{"Assigning", "my-user", "my-space", "SpaceManager"},
		))

		Expect(spaceRepo.CreateCallCount()).To(Equal(1))
		actualSpaceName, _, _ := spaceRepo.CreateArgsForCall(0)
		Expect(actualSpaceName).To(Equal("my-space"))
		Expect(userRepo.SetSpaceRoleByGUIDCallCount()).To(BeZero())
	})

	Context("when the -o flag is provided", func() {
		It("creates a space within that org", func() {
			org := models.Organization{
				OrganizationFields: models.OrganizationFields{
					Name: "other-org",
					GUID: "org-guid-1",
				}}
			orgRepo.FindByNameReturns(org, nil)

			runCommand("-o", "other-org", "my-space")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating space", "my-space", "other-org", "my-user"},
				[]string{"OK"},
				[]string{"Assigning", "my-user", "my-space", "SpaceManager"},
				[]string{"Assigning", "my-user", "my-space", "SpaceDeveloper"},
				[]string{"TIP"},
			))

			actualSpaceName, actualOrgGUID, _ := spaceRepo.CreateArgsForCall(0)
			Expect(actualSpaceName).To(Equal("my-space"))
			Expect(actualOrgGUID).To(Equal(org.GUID))

			userGUID, spaceGUID, orgGUID, role := userRepo.SetSpaceRoleByGUIDArgsForCall(0)
			Expect(userGUID).To(Equal("my-user-guid"))
			Expect(spaceGUID).To(Equal("my-space-guid"))
			Expect(orgGUID).To(Equal("org-guid-1"))
			Expect(role).To(Equal(models.RoleSpaceManager))
		})

		It("fails when the org provided does not exist", func() {
			orgRepo.FindByNameReturns(models.Organization{}, errors.New("cool-organization does not exist"))
			runCommand("-o", "cool-organization", "my-space")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"cool-organization", "does not exist"},
			))

			Expect(spaceRepo.CreateCallCount()).To(BeZero())
		})

		It("fails when finding the org returns an error", func() {
			orgRepo.FindByNameReturns(models.Organization{}, errors.New("cool-organization does not exist"))
			runCommand("-o", "cool-organization", "my-space")

			Expect(ui.Outputs()).To(ContainSubstrings(
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
				GUID: "my-space-quota-guid",
			}
			spaceQuotaRepo.FindByNameAndOrgGUIDReturns(spaceQuota, nil)
			runCommand("-q", "my-space-quota", "my-space")

			spaceQuotaName, orgGUID := spaceQuotaRepo.FindByNameAndOrgGUIDArgsForCall(0)

			Expect(spaceQuotaName).To(Equal(spaceQuota.Name))
			Expect(orgGUID).To(Equal(configOrg.GUID))

			_, _, actualSpaceQuotaGUID := spaceRepo.CreateArgsForCall(0)
			Expect(actualSpaceQuotaGUID).To(Equal(spaceQuota.GUID))
		})

		Context("when the space-quota provided does not exist", func() {
			It("fails", func() {
				spaceQuotaRepo.FindByNameAndOrgGUIDReturns(models.SpaceQuota{}, errors.New("Error"))
				runCommand("-q", "my-space-quota", "my-space")

				Expect(ui.Outputs()).To(ContainSubstrings(
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
					GUID: "other-org-guid",
				}}
			orgRepo.FindByNameReturns(org, nil)

			spaceQuota := models.SpaceQuota{
				Name: "my-space-quota",
				GUID: "my-space-quota-guid",
			}
			spaceQuotaRepo.FindByNameAndOrgGUIDReturns(spaceQuota, nil)
		})

		It("assigns the space-quota from the specified org to the space", func() {
			runCommand("my-space", "-o", "other-org", "-q", "my-space-quota")

			Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("other-org"))
			spaceName, orgGUID := spaceQuotaRepo.FindByNameAndOrgGUIDArgsForCall(0)
			Expect(spaceName).To(Equal("my-space-quota"))
			Expect(orgGUID).To(Equal("other-org-guid"))

			actualSpaceName, actualOrgGUID, actualSpaceQuotaGUID := spaceRepo.CreateArgsForCall(0)
			Expect(actualSpaceName).To(Equal("my-space"))
			Expect(actualOrgGUID).To(Equal("other-org-guid"))
			Expect(actualSpaceQuotaGUID).To(Equal("my-space-quota-guid"))
		})
	})
})
