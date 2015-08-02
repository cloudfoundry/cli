package service_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
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
		config              core_config.Repository
		repo                *testapi.FakeUserProvidedServiceInstanceRepository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetUserProvidedServiceInstanceRepository(repo)
		deps.Config = config
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("create-user-provided-service").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		repo = &testapi.FakeUserProvidedServiceInstanceRepository{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	})

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(testcmd.RunCliCommand("create-user-provided-service", []string{"my-service"}, requirementsFactory, updateCommandDependency, false)).To(BeFalse())
		})
		It("fails when a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(testcmd.RunCliCommand("create-user-provided-service", []string{"my-service"}, requirementsFactory, updateCommandDependency, false)).To(BeFalse())
		})
	})

	It("creates a new user provided service given just a name", func() {
		testcmd.RunCliCommand("create-user-provided-service", []string{"my-custom-service"}, requirementsFactory, updateCommandDependency, false)
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user provided service"},
			[]string{"OK"},
		))
	})

	It("accepts service parameters interactively", func() {
		ui.Inputs = []string{"foo value", "bar value", "baz value"}
		testcmd.RunCliCommand("create-user-provided-service", []string{"-p", `"foo, bar, baz"`, "my-custom-service"}, requirementsFactory, updateCommandDependency, false)

		Expect(ui.Prompts).To(ContainSubstrings(
			[]string{"foo"},
			[]string{"bar"},
			[]string{"baz"},
		))

		Expect(repo.CreateCallCount()).To(Equal(1))
		name, drainUrl, params := repo.CreateArgsForCall(0)
		Expect(name).To(Equal("my-custom-service"))
		Expect(drainUrl).To(Equal(""))
		Expect(params["foo"]).To(Equal("foo value"))
		Expect(params["bar"]).To(Equal("bar value"))
		Expect(params["baz"]).To(Equal("baz value"))

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user provided service", "my-custom-service", "my-org", "my-space", "my-user"},
			[]string{"OK"},
		))
	})

	It("accepts service parameters as JSON without prompting", func() {
		args := []string{"-p", `{"foo": "foo value", "bar": "bar value", "baz": 4}`, "my-custom-service"}
		testcmd.RunCliCommand("create-user-provided-service", args, requirementsFactory, updateCommandDependency, false)

		name, _, params := repo.CreateArgsForCall(0)
		Expect(name).To(Equal("my-custom-service"))

		Expect(ui.Prompts).To(BeEmpty())
		Expect(params).To(Equal(map[string]interface{}{
			"foo": "foo value",
			"bar": "bar value",
			"baz": float64(4),
		}))

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user provided service"},
			[]string{"OK"},
		))
	})

	It("creates a user provided service with a syslog drain url", func() {
		args := []string{"-l", "syslog://example.com", "-p", `{"foo": "foo value", "bar": "bar value", "baz": "baz value"}`, "my-custom-service"}
		testcmd.RunCliCommand("create-user-provided-service", args, requirementsFactory, updateCommandDependency, false)

		_, drainUrl, _ := repo.CreateArgsForCall(0)
		Expect(drainUrl).To(Equal("syslog://example.com"))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user provided service"},
			[]string{"OK"},
		))
	})
})
