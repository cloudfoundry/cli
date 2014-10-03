package user_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/user"
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

var _ = Describe("space-users command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		spaceRepo           *testapi.FakeSpaceRepository
		userRepo            *testapi.FakeUserRepository
		config              core_config.ReadWriter
	)

	BeforeEach(func() {
		config = testconfig.NewRepositoryWithDefaults()
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		spaceRepo = &testapi.FakeSpaceRepository{}
		userRepo = &testapi.FakeUserRepository{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(NewSpaceUsers(ui, config, spaceRepo, userRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			runCommand("my-org", "my-space")

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("succeeds when logged in", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("some-org", "whatever-space")

			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
			Expect("some-org").To(Equal(requirementsFactory.OrganizationName))
		})
	})

	It("fails with usage when not invoked with exactly two args", func() {
		runCommand("my-org")
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	Context("when logged in and given some users in the org and space", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true

			org := models.Organization{}
			org.Name = "Org1"
			org.Guid = "org1-guid"
			space := models.Space{}
			space.Name = "Space1"
			space.Guid = "space1-guid"

			requirementsFactory.Organization = org
			spaceRepo.FindByNameInOrgSpace = space

			user := models.UserFields{}
			user.Username = "user1"
			user2 := models.UserFields{}
			user2.Username = "user2"
			user3 := models.UserFields{}
			user3.Username = "user3"
			user4 := models.UserFields{}
			user4.Username = "user4"
			userRepo.ListUsersByRole = map[string][]models.UserFields{
				models.SPACE_MANAGER:   []models.UserFields{user, user2},
				models.SPACE_DEVELOPER: []models.UserFields{user4},
				models.SPACE_AUDITOR:   []models.UserFields{user3},
			}
		})

		It("tells you about the space users in the given space", func() {
			runCommand("my-org", "my-space")

			Expect(spaceRepo.FindByNameInOrgName).To(Equal("my-space"))
			Expect(spaceRepo.FindByNameInOrgOrgGuid).To(Equal("org1-guid"))
			Expect(userRepo.ListUsersSpaceGuid).To(Equal("space1-guid"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting users in org", "Org1", "Space1", "my-user"},
				[]string{"SPACE MANAGER"},
				[]string{"user1"},
				[]string{"user2"},
				[]string{"SPACE DEVELOPER"},
				[]string{"user4"},
				[]string{"SPACE AUDITOR"},
				[]string{"user3"},
			))
		})
	})
})
