package commands_test

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("config command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("config").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCliCommand("config", args, requirementsFactory, updateCommandDependency, false)
	}
	It("fails requirements when no flags are provided", func() {
		runCommand()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage"},
		))
	})

	Context("--async-timeout flag", func() {

		It("stores the timeout in minutes when the --async-timeout flag is provided", func() {
			runCommand("--async-timeout", "12")
			Expect(configRepo.AsyncTimeout()).Should(Equal(uint(12)))
		})

		It("fails with usage when a invalid async timeout value is passed", func() {
			runCommand("--async-timeout", "-1")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage"},
			))
		})

		It("fails with usage when a negative timout is passed", func() {
			runCommand("--async-timeout", "-555")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage"},
			))
			Expect(configRepo.AsyncTimeout()).To(Equal(uint(0)))
		})
	})

	Context("--trace flag", func() {
		It("stores the trace value when --trace flag is provided", func() {
			runCommand("--trace", "true")
			Expect(configRepo.Trace()).Should(Equal("true"))

			runCommand("--trace", "false")
			Expect(configRepo.Trace()).Should(Equal("false"))

			runCommand("--trace", "some/file/lol")
			Expect(configRepo.Trace()).Should(Equal("some/file/lol"))
		})
	})

	Context("--color flag", func() {
		It("stores the color value when --color flag is provided", func() {
			runCommand("--color", "true")
			Expect(configRepo.ColorEnabled()).Should(Equal("true"))

			runCommand("--color", "false")
			Expect(configRepo.ColorEnabled()).Should(Equal("false"))
		})

		It("fails with usage when a non-bool value is provided", func() {
			runCommand("--color", "plaid")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage"},
			))
		})
	})

	Context("--locale flag", func() {
		It("stores the locale value when --locale [locale] is provided", func() {
			runCommand("--locale", "zh_Hans")
			Expect(configRepo.Locale()).Should(Equal("zh_Hans"))
		})

		It("informs the user of known locales if an unknown locale is provided", func() {
			runCommand("--locale", "foo_BAR")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"known locales are"},
				[]string{"en_US"},
				[]string{"fr_FR"},
			))
		})

		Context("when the locale is already set", func() {
			BeforeEach(func() {
				configRepo.SetLocale("fr_FR")
				Expect(configRepo.Locale()).Should(Equal("fr_FR"))
			})

			It("clears the locale when the '--locale clear' flag is provided", func() {
				runCommand("--locale", "CLEAR")
				Expect(configRepo.Locale()).Should(Equal(""))
			})
		})
	})
})
