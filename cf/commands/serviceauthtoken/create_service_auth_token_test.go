package serviceauthtoken_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/serviceauthtoken"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-service-auth-token command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.ReadWriter
		authTokenRepo       *testapi.FakeAuthTokenRepo
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		authTokenRepo = &testapi.FakeAuthTokenRepo{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		cmd := NewCreateServiceAuthToken(ui, configRepo, authTokenRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when not invoked with exactly three args", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("whoops", "i-accidentally-an-arg")

			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			runCommand("just", "enough", "args")

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("creates a service auth token, obviously", func() {
			runCommand("a label", "a provider", "a value")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service auth token as", "my-user"},
				[]string{"OK"},
			))

			authToken := models.ServiceAuthTokenFields{}
			authToken.Label = "a label"
			authToken.Provider = "a provider"
			authToken.Token = "a value"
			Expect(authTokenRepo.CreatedServiceAuthTokenFields).To(Equal(authToken))
		})
	})
})
