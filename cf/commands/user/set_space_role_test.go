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

var _ = Describe("set-space-role command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		spaceRepo           *testapi.FakeSpaceRepository
		userRepo            *testapi.FakeUserRepository
		configRepo          core_config.ReadWriter
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		accessToken, err := testconfig.EncodeAccessToken(core_config.TokenInfo{Username: "current-user"})
		Expect(err).NotTo(HaveOccurred())
		configRepo.SetAccessToken(accessToken)

		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		spaceRepo = &testapi.FakeSpaceRepository{}
		userRepo = &testapi.FakeUserRepository{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewSetSpaceRole(ui, configRepo, spaceRepo, userRepo), args, requirementsFactory)
	}

	It("fails with usage when not provided exactly four args", func() {
		runCommand("foo", "bar", "baz")
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("does not fail with usage when provided four args", func() {
		runCommand("whatever", "these", "are", "args")
		Expect(ui.FailedWithUsage).To(BeFalse())
	})

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			Expect(runCommand("username", "org", "space", "role")).To(BeFalse())
		})

		It("succeeds when logged in", func() {
			requirementsFactory.LoginSuccess = true
			passed := runCommand("username", "org", "space", "role")

			Expect(passed).To(BeTrue())
			Expect(requirementsFactory.UserUsername).To(Equal("username"))
			Expect(requirementsFactory.OrganizationName).To(Equal("org"))
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true

			org := models.Organization{}
			org.Guid = "my-org-guid"
			org.Name = "my-org"

			requirementsFactory.UserFields = models.UserFields{Guid: "my-user-guid", Username: "my-user"}
			requirementsFactory.Organization = org

			spaceRepo.FindByNameInOrgSpace = models.Space{}
			spaceRepo.FindByNameInOrgSpace.Guid = "my-space-guid"
			spaceRepo.FindByNameInOrgSpace.Name = "my-space"
			spaceRepo.FindByNameInOrgSpace.Organization = org.OrganizationFields
		})

		It("sets the given space role on the given user", func() {
			runCommand("some-user", "some-org", "some-space", "SpaceManager")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Assigning role ", "SpaceManager", "my-user", "my-org", "my-space", "current-user"},
				[]string{"OK"},
			))

			Expect(spaceRepo.FindByNameInOrgName).To(Equal("some-space"))
			Expect(spaceRepo.FindByNameInOrgOrgGuid).To(Equal("my-org-guid"))

			Expect(userRepo.SetSpaceRoleUserGuid).To(Equal("my-user-guid"))
			Expect(userRepo.SetSpaceRoleSpaceGuid).To(Equal("my-space-guid"))
			Expect(userRepo.SetSpaceRoleRole).To(Equal(models.SPACE_MANAGER))
		})
	})
})
