package user_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("unset-space-role command", func() {

	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		userRepo            *testapi.FakeUserRepository
		spaceRepo           *testapi.FakeSpaceRepository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetUserRepository(userRepo)
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("unset-space-role").SetDependency(deps, pluginCall))
	}

	callUnsetSpaceRole := func(args []string, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository, requirementsFactory *testreq.FakeReqFactory) (*testterm.FakeUI, bool) {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		passed := testcmd.RunCliCommand("unset-space-role", args, requirementsFactory, updateCommandDependency, false)
		return ui, passed
	}
	It("fails with usage when not called with exactly four args", func() {
		requirementsFactory, spaceRepo, userRepo = getUnsetSpaceRoleDeps()

		ui, _ := callUnsetSpaceRole([]string{"username", "org", "space"}, spaceRepo, userRepo, requirementsFactory)
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))
	})

	It("fails requirements when not logged in", func() {
		requirementsFactory, spaceRepo, userRepo = getUnsetSpaceRoleDeps()
		args := []string{"username", "org", "space", "role"}

		requirementsFactory.LoginSuccess = false
		_, passed := callUnsetSpaceRole(args, spaceRepo, userRepo, requirementsFactory)
		Expect(passed).To(BeFalse())
	})

	It("unsets the user's space role", func() {
		user := models.UserFields{}
		user.Username = "some-user"
		user.Guid = "some-user-guid"
		org := models.Organization{}
		org.Name = "some-org"
		org.Guid = "some-org-guid"

		requirementsFactory, spaceRepo, userRepo = getUnsetSpaceRoleDeps()
		requirementsFactory.LoginSuccess = true
		requirementsFactory.UserFields = user
		requirementsFactory.Organization = org
		spaceRepo.FindByNameInOrgSpace = models.Space{}
		spaceRepo.FindByNameInOrgSpace.Name = "some-space"
		spaceRepo.FindByNameInOrgSpace.Guid = "some-space-guid"

		args := []string{"my-username", "my-org", "my-space", "SpaceManager"}

		ui, _ := callUnsetSpaceRole(args, spaceRepo, userRepo, requirementsFactory)

		Expect(spaceRepo.FindByNameInOrgName).To(Equal("my-space"))
		Expect(spaceRepo.FindByNameInOrgOrgGuid).To(Equal("some-org-guid"))

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Removing role", "SpaceManager", "some-user", "some-org", "some-space", "my-user"},
			[]string{"OK"},
		))
		Expect(userRepo.UnsetSpaceRoleRole).To(Equal(models.SPACE_MANAGER))
		Expect(userRepo.UnsetSpaceRoleUserGuid).To(Equal("some-user-guid"))
		Expect(userRepo.UnsetSpaceRoleSpaceGuid).To(Equal("some-space-guid"))
	})
})

func getUnsetSpaceRoleDeps() (requirementsFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) {
	requirementsFactory = &testreq.FakeReqFactory{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	userRepo = &testapi.FakeUserRepository{}
	return
}
