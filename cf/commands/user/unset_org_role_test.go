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

var _ = Describe("unset-org-role command", func() {
	var (
		ui                  *testterm.FakeUI
		userRepo            *testapi.FakeUserRepository
		configRepo          core_config.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
	)

	runCommand := func(args ...string) bool {
		cmd := NewUnsetOrgRole(ui, configRepo, userRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		userRepo = &testapi.FakeUserRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	It("fails with usage when invoked without exactly three args", func() {
		runCommand("username", "org")
		Expect(ui.FailedWithUsage).To(BeTrue())

		runCommand("woah", "too", "many", "args")
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false

			Expect(runCommand("username", "org", "role")).To(BeFalse())
		})

		It("succeeds when logged in", func() {
			requirementsFactory.LoginSuccess = true
			passed := runCommand("username", "org", "role")
			Expect(passed).To(BeTrue())

			Expect(requirementsFactory.UserUsername).To(Equal("username"))
			Expect(requirementsFactory.OrganizationName).To(Equal("org"))
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true

			user := models.UserFields{}
			user.Username = "some-user"
			user.Guid = "some-user-guid"
			org := models.Organization{}
			org.Name = "some-org"
			org.Guid = "some-org-guid"

			requirementsFactory.UserFields = user
			requirementsFactory.Organization = org
		})

		It("unsets a user's org role", func() {
			runCommand("my-username", "my-org", "OrgManager")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Removing role", "OrgManager", "my-username", "my-org", "my-user"},
				[]string{"OK"},
			))

			Expect(userRepo.UnsetOrgRoleRole).To(Equal(models.ORG_MANAGER))
			Expect(userRepo.UnsetOrgRoleUserGuid).To(Equal("some-user-guid"))
			Expect(userRepo.UnsetOrgRoleOrganizationGuid).To(Equal("some-org-guid"))
		})
	})
})
