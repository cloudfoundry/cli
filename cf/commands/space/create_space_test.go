package space_test

import (
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
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
		configRepo          configuration.ReadWriter
		spaceRepo           *testapi.FakeSpaceRepository
		orgRepo             *testapi.FakeOrgRepository
		userRepo            *testapi.FakeUserRepository
		spaceRoleSetter     user.SpaceRoleSetter
	)

	runCommand := func(args ...string) {
		cmd := NewCreateSpace(ui, configRepo, spaceRoleSetter, spaceRepo, orgRepo, userRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()

		orgRepo = &testapi.FakeOrgRepository{}
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
		It("fails with usage when no arguments are passed", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				requirementsFactory.LoginSuccess = false
			})

			It("fails requirements", func() {
				runCommand("some-space")
				Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
			})
		})

		Context("when a org is not targeted", func() {
			BeforeEach(func() {
				requirementsFactory.TargetedOrgSuccess = false
			})

			It("fails requirements", func() {
				runCommand("what-is-space?")
				Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
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
			org := maker.NewOrg(maker.Overrides{"name": "other-org"})
			orgRepo.Organizations = []models.Organization{org}

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
			orgRepo.FindByNameNotFound = true
			runCommand("-o", "cool-organization", "my-space")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"cool-organization", "does not exist"},
			))

			Expect(spaceRepo.CreateSpaceName).To(Equal(""))
		})

		It("fails when finding the org returns an error", func() {
			orgRepo.FindByNameErr = true
			runCommand("-o", "cool-organization", "my-space")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"Error"},
			))

			Expect(spaceRepo.CreateSpaceName).To(Equal(""))
		})
	})
})
