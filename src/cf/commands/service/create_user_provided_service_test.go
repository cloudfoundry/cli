package service_test

import (
	. "cf/commands/service"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
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
			ctxt := testcmd.NewContext("create-user-provided-service", []string{"my-service"})
			testcmd.RunCommand(cmd, ctxt, requirementsFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	It("creates a new user provided service given just a name", func() {
		args := []string{"my-custom-service"}
		ctxt := testcmd.NewContext("create-user-provided-service", args)
		testcmd.RunCommand(cmd, ctxt, requirementsFactory)
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating user provided service"},
			{"OK"},
		})
	})

	It("accepts service parameters interactively", func() {
		ui.Inputs = []string{"foo value", "bar value", "baz value"}
		ctxt := testcmd.NewContext("create-user-provided-service", []string{"-p", `"foo, bar, baz"`, "my-custom-service"})
		testcmd.RunCommand(cmd, ctxt, requirementsFactory)

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"foo"},
			{"bar"},
			{"baz"},
		})

		Expect(repo.CreateName).To(Equal("my-custom-service"))
		Expect(repo.CreateParams).To(Equal(map[string]string{
			"foo": "foo value",
			"bar": "bar value",
			"baz": "baz value",
		}))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating user provided service", "my-custom-service", "my-org", "my-space", "my-user"},
			{"OK"},
		})
	})

	It("accepts service parameters as JSON without prompting", func() {
		args := []string{"-p", `{"foo": "foo value", "bar": "bar value", "baz": "baz value"}`, "my-custom-service"}
		ctxt := testcmd.NewContext("create-user-provided-service", args)
		testcmd.RunCommand(cmd, ctxt, requirementsFactory)

		Expect(ui.Prompts).To(BeEmpty())
		Expect(repo.CreateName).To(Equal("my-custom-service"))
		Expect(repo.CreateParams).To(Equal(map[string]string{
			"foo": "foo value",
			"bar": "bar value",
			"baz": "baz value",
		}))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating user provided service"},
			{"OK"},
		})
	})

	It("creates a user provided service with a syslog drain url", func() {
		args := []string{"-l", "syslog://example.com", "-p", `{"foo": "foo value", "bar": "bar value", "baz": "baz value"}`, "my-custom-service"}
		ctxt := testcmd.NewContext("create-user-provided-service", args)
		testcmd.RunCommand(cmd, ctxt, requirementsFactory)

		Expect(repo.CreateDrainUrl).To(Equal("syslog://example.com"))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating user provided service"},
			{"OK"},
		})
	})
})
