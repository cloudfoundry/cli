package organization_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/organization"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-org command", func() {
	var (
		config              configuration.ReadWriter
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		orgRepo             *testapi.FakeOrganizationRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		orgRepo = &testapi.FakeOrganizationRepository{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(NewCreateOrg(ui, config, orgRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when not provided exactly one arg", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			runCommand("my-org")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when logged in and provided the name of an org to create", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("creates an org", func() {
			runCommand("my-org")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating org", "my-org", "my-user"},
				[]string{"OK"},
			))
			Expect(orgRepo.CreateName).To(Equal("my-org"))
		})

		It("fails and warns the user when the org already exists", func() {
			orgRepo.CreateOrgExists = true
			runCommand("my-org")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating org", "my-org"},
				[]string{"OK"},
				[]string{"my-org", "already exists"},
			))
		})
	})
})
