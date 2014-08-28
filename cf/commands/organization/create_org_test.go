package organization_test

import (
	"github.com/cloudfoundry/cli/cf/errors"

	test_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
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
		orgRepo             *test_org.FakeOrganizationRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		orgRepo = &test_org.FakeOrganizationRepository{}
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
			orgRepo.CreateReturns(nil)
			requirementsFactory.LoginSuccess = true
		})

		It("creates an org", func() {
			runCommand("my-org")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating org", "my-org", "my-user"},
				[]string{"OK"},
			))
			Expect(orgRepo.CreateArgsForCall(0)).To(Equal("my-org"))
		})

		It("fails and warns the user when the org already exists", func() {
			err := errors.NewHttpError(400, errors.ORG_EXISTS, "org already exists")
			orgRepo.CreateReturns(err)
			runCommand("my-org")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating org", "my-org"},
				[]string{"OK"},
				[]string{"my-org", "already exists"},
			))
		})
	})
})
