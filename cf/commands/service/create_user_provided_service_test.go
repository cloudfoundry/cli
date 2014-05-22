package service_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/configuration"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("create-user-provided-service command", func() {
	var (
		ui                  *testterm.FakeUI
		config              configuration.ReadWriter
		repo                *testapi.FakeUserProvidedServiceInstanceRepo
		requirementsFactory *testreq.FakeReqFactory
		cmd                 CreateUserProvidedService
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		repo = &testapi.FakeUserProvidedServiceInstanceRepo{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		cmd = NewCreateUserProvidedService(ui, config, repo)
	})

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			testcmd.RunCommand(cmd, []string{"my-service"}, requirementsFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	It("creates a new user provided service given just a name", func() {
		testcmd.RunCommand(cmd, []string{"my-custom-service"}, requirementsFactory)
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user provided service"},
			[]string{"OK"},
		))
	})

	It("accepts service parameters interactively", func() {
		ui.Inputs = []string{"foo value", "bar value", "baz value"}
		testcmd.RunCommand(cmd, []string{"-p", `"foo, bar, baz"`, "my-custom-service"}, requirementsFactory)

		Expect(ui.Prompts).To(ContainSubstrings(
			[]string{"foo"},
			[]string{"bar"},
			[]string{"baz"},
		))

		Expect(repo.CreateName).To(Equal("my-custom-service"))
		Expect(repo.CreateParams).To(Equal(map[string]string{
			"foo": "foo value",
			"bar": "bar value",
			"baz": "baz value",
		}))

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user provided service", "my-custom-service", "my-org", "my-space", "my-user"},
			[]string{"OK"},
		))
	})

	It("accepts service parameters as JSON without prompting", func() {
		args := []string{"-p", `{"foo": "foo value", "bar": "bar value", "baz": "baz value"}`, "my-custom-service"}
		testcmd.RunCommand(cmd, args, requirementsFactory)

		Expect(ui.Prompts).To(BeEmpty())
		Expect(repo.CreateName).To(Equal("my-custom-service"))
		Expect(repo.CreateParams).To(Equal(map[string]string{
			"foo": "foo value",
			"bar": "bar value",
			"baz": "baz value",
		}))

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user provided service"},
			[]string{"OK"},
		))
	})

	It("creates a user provided service with a syslog drain url", func() {
		args := []string{"-l", "syslog://example.com", "-p", `{"foo": "foo value", "bar": "bar value", "baz": "baz value"}`, "my-custom-service"}
		testcmd.RunCommand(cmd, args, requirementsFactory)

		Expect(repo.CreateDrainUrl).To(Equal("syslog://example.com"))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user provided service"},
			[]string{"OK"},
		))
	})
})
