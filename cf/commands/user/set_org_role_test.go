package user_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

func callSetOrgRole(args []string, requirementsFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	config := testconfig.NewRepositoryWithDefaults()
	cmd := NewSetOrgRole(ui, config, userRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}

var _ = Describe("set-org-role command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		userRepo            *testapi.FakeUserRepository
		configRepo          configuration.ReadWriter
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()

		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		userRepo = &testapi.FakeUserRepository{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(NewSetOrgRole(ui, configRepo, userRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			runCommand("hey", "there", "jude")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails with usage when not provided exactly three args", func() {
			runCommand("one fish", "two fish") // red fish, blue fish
			Expect(ui.FailedWithUsage).To(BeTrue())
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
		})

		It("sets the given org role on the given user", func() {
			runCommand("some-user", "some-org", "OrgManager")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Assigning role", "OrgManager", "my-user", "my-org", "my-user"},
				[]string{"OK"},
			))
			Expect(userRepo.SetOrgRoleUserGuid).To(Equal("my-user-guid"))
			Expect(userRepo.SetOrgRoleOrganizationGuid).To(Equal("my-org-guid"))
			Expect(userRepo.SetOrgRoleRole).To(Equal(models.ORG_MANAGER))
		})
	})
})
