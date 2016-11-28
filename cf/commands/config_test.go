package commands_test

import (
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("config command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("config").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
	})

	runCommand := func(args ...string) {
		testcmd.RunCLICommand("config", args, requirementsFactory, updateCommandDependency, false, ui)
	}
	It("fails requirements when no flags are provided", func() {
		runCommand()
		Expect(ui.Outputs()).To(ContainSubstrings(
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
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage"},
			))
		})

		It("fails with usage when a negative timout is passed", func() {
			runCommand("--async-timeout", "-555")
			Expect(ui.Outputs()).To(ContainSubstrings(
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
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage"},
			))
		})
	})

	Context("--locale flag", func() {
		It("stores the locale value when --locale [locale] is provided", func() {
			runCommand("--locale", "zh-Hans")
			Expect(configRepo.Locale()).Should(Equal("zh-Hans"))
		})

		It("informs the user of known locales if an unknown locale is provided", func() {
			runCommand("--locale", "foo-BAR")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"Could not find locale 'foo-BAR'. The known locales are:"},
				[]string{"en-US"},
				[]string{"fr-FR"},
				[]string{"zh-Hans"},
			))
		})

		Context("when the locale is already set", func() {
			BeforeEach(func() {
				configRepo.SetLocale("fr-FR")
				Expect(configRepo.Locale()).Should(Equal("fr-FR"))
			})

			It("clears the locale when the '--locale clear' flag is provided", func() {
				runCommand("--locale", "CLEAR")
				Expect(configRepo.Locale()).Should(Equal(""))
			})
		})
	})
})
