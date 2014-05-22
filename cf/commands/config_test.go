package commands_test

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("config command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		cmd := NewConfig(ui, configRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}
	It("fails requirements when no flags are provided", func() {
		runCommand()
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	Context("--async-timeout flag", func() {

		It("stores the timeout in minutes when the --async-timeout flag is provided", func() {
			runCommand("--async-timeout", "12")
			Expect(configRepo.AsyncTimeout()).Should(Equal(uint(12)))
		})

		It("fails with usage when a invalid async timeout value is passed, e.g., a string", func() {
			runCommand("--async-timeout", "lol")
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails with usage when a negative timout is passed", func() {
			runCommand("--async-timeout", "-555")
			Expect(ui.FailedWithUsage).To(BeTrue())
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
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})
})
