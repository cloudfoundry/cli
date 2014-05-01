package serviceauthtoken_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/serviceauthtoken"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("delete-service-auth-token command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
		authTokenRepo       *testapi.FakeAuthTokenRepo
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{Inputs: []string{"y"}}
		authTokenRepo = &testapi.FakeAuthTokenRepo{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	})

	runCommand := func(args ...string) {
		cmd := NewDeleteServiceAuthToken(ui, configRepo, authTokenRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-service-auth-token", args), requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when fewer than two arguments are given", func() {
			runCommand("yurp")
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			runCommand()
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when the service auth token exists", func() {
		BeforeEach(func() {
			authTokenRepo.FindByLabelAndProviderServiceAuthTokenFields = models.ServiceAuthTokenFields{
				Guid:     "the-guid",
				Label:    "a label",
				Provider: "a provider",
			}
		})

		It("deletes the service auth token", func() {
			runCommand("a label", "a provider")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting service auth token as", "my-user"},
				[]string{"OK"},
			))

			Expect(authTokenRepo.FindByLabelAndProviderLabel).To(Equal("a label"))
			Expect(authTokenRepo.FindByLabelAndProviderProvider).To(Equal("a provider"))
			Expect(authTokenRepo.DeletedServiceAuthTokenFields.Guid).To(Equal("the-guid"))
		})

		It("does nothing when the user does not confirm", func() {
			ui.Inputs = []string{"nope"}
			runCommand("a label", "a provider")

			Expect(ui.Prompts).To(ContainSubstrings(
				[]string{"Really delete", "service auth token", "a label", "a provider"},
			))
			Expect(ui.Outputs).To(BeEmpty())
			Expect(authTokenRepo.DeletedServiceAuthTokenFields).To(Equal(models.ServiceAuthTokenFields{}))
		})

		It("does not prompt the user when the -f flag is given", func() {
			ui.Inputs = []string{}
			runCommand("-f", "a label", "a provider")

			Expect(ui.Prompts).To(BeEmpty())
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting"},
				[]string{"OK"},
			))

			Expect(authTokenRepo.DeletedServiceAuthTokenFields.Guid).To(Equal("the-guid"))
		})
	})

	Context("when the service auth token does not exist", func() {
		BeforeEach(func() {
			authTokenRepo.FindByLabelAndProviderApiResponse = errors.NewModelNotFoundError("Service Auth Token", "")
		})

		It("warns the user when the specified service auth token does not exist", func() {
			runCommand("a label", "a provider")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting service auth token as", "my-user"},
				[]string{"OK"},
			))

			Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"does not exist"}))
		})
	})

	Context("when there is an error deleting the service auth token", func() {
		BeforeEach(func() {
			authTokenRepo.FindByLabelAndProviderApiResponse = errors.New("OH NOES")
		})

		It("TestDeleteServiceAuthTokenFailsWithError", func() {
			runCommand("a label", "a provider")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting service auth token as", "my-user"},
				[]string{"FAILED"},
				[]string{"OH NOES"},
			))
		})
	})
})
