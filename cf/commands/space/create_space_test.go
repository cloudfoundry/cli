package space_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fake_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	quotafakes "github.com/cloudfoundry/cli/cf/api/space_quotas/fakes"
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/testhelpers/maker"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/space"
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
		configRepo          core_config.ReadWriter
		spaceRepo           *testapi.FakeSpaceRepository
		orgRepo             *fake_org.FakeOrganizationRepository
		userRepo            *testapi.FakeUserRepository
		spaceRoleSetter     user.SpaceRoleSetter
		spaceQuotaRepo      *quotafakes.FakeSpaceQuotaRepository
	)

	runCommand := func(args ...string) bool {
		cmd := NewCreateSpace(ui, configRepo, spaceRoleSetter, spaceRepo, orgRepo, userRepo, spaceQuotaRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()

		orgRepo = &fake_org.FakeOrganizationRepository{}
		spaceQuotaRepo = &quotafakes.FakeSpaceQuotaRepository{}
		userRepo = &testapi.FakeUserRepository{}
		spaceRoleSetter = user.NewSetSpaceRole(ui, configRepo, spaceRepo, userRepo)

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		configOrg = models.OrganizationFields{
			Name: "my-org",
			Guid: "my-org-guid",
		}

		configSpace = models.SpaceFields{
			Name: "config-space",
			Guid: "config-space-guid",
		}

		spaceRepo = &testapi.FakeSpaceRepository{
			CreateSpaceSpace: maker.NewSpace(maker.Overrides{"name": "my-space", "guid": "my-space-guid", "organization": configOrg}),
		}
		Expect(spaceRepo.CreateSpaceSpace.Name).To(Equal("my-space"))
	})

	Describe("Requirements", func() {
		It("fails with usage when not provided exactly one argument", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
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
		It("creates a space and associate with the provided space-quota", func() {
			spaceQuotaRepo.FindByNameReturns(
				models.SpaceQuota{
					Name:                    "quota-name",
					Guid:                    "quota-guid",
					MemoryLimit:             1024,
					InstanceMemoryLimit:     512,
					RoutesLimit:             111,
					ServicesLimit:           222,
					NonBasicServicesAllowed: true,
					OrgGuid:                 "my-org-guid",
				}, nil)

			runCommand("-q", "quota-name", "my-space")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating space", "my-space", "my-org", "my-user"},
				[]string{"OK"},
				[]string{"Assigning", models.SpaceRoleToUserInput[models.SPACE_MANAGER], "my-user", "my-space"},
				[]string{"Assigning", models.SpaceRoleToUserInput[models.SPACE_DEVELOPER], "my-user", "my-space"},
				[]string{"Assigning space quota", "to space", "my-user"},
				[]string{"TIP"},
			))

			Expect(spaceRepo.CreateSpaceName).To(Equal("my-space"))
			Expect(spaceRepo.CreateSpaceOrgGuid).To(Equal("my-org-guid"))
			Expect(userRepo.SetSpaceRoleUserGuid).To(Equal("my-user-guid"))
			Expect(userRepo.SetSpaceRoleSpaceGuid).To(Equal("my-space-guid"))
			Expect(userRepo.SetSpaceRoleRole).To(Equal(models.SPACE_DEVELOPER))
		})

		It("when an error occurs fetching the space-quota", func() {
			spaceQuotaRepo.FindByNameReturns(models.SpaceQuota{}, errors.New("Space Quota unknown-quota-name not found"))

			runCommand("-q", "unknown-quota-name", "my-space")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating space", "my-space", "my-org", "my-user"},
				[]string{"FAILED"},
				[]string{"Space Quota", "not found"},
			))
		})
	})
})
