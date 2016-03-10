package rpc_test

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/plugin/rpc"
	. "github.com/cloudfoundry/cli/plugin/rpc/fake_command"

	"github.com/cloudfoundry/cli/cf/trace/fakes"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("calling commands in command_registry", func() {

	_ = FakeCommand1{} //make sure fake_command is imported and self-registered with init()

	var (
		ui         *testterm.FakeUI
		deps       command_registry.Dependency
		fakeLogger *fakes.FakePrinter
	)

	BeforeEach(func() {
		fakeLogger = new(fakes.FakePrinter)
		deps = command_registry.NewDependency(fakeLogger)
		ui = &testterm.FakeUI{}
		deps.Ui = ui

		cmd := command_registry.Commands.FindCommand("fake-command")
		command_registry.Commands.SetCommand(cmd.SetDependency(deps, true))

		cmd2 := command_registry.Commands.FindCommand("fake-command2")
		command_registry.Commands.SetCommand(cmd2.SetDependency(deps, true))
	})

	It("runs the command requirements", func() {
		NewCommandRunner().Command([]string{"fake-command"}, deps, false)
		Expect(ui.Outputs).To(ContainSubstrings([]string{"Requirement executed"}))
	})

	It("calls the command Execute() func", func() {
		NewCommandRunner().Command([]string{"fake-command"}, deps, false)
		Expect(ui.Outputs).To(ContainSubstrings([]string{"Command Executed"}))
	})

	It("sets the dependency of the command", func() {
		NewCommandRunner().Command([]string{"fake-command"}, deps, false)
		Expect(ui.Outputs).To(ContainSubstrings([]string{"SetDependency() called, pluginCall true"}))
	})

	It("returns an error if any of the requirements fail", func() {
		err := NewCommandRunner().Command([]string{"fake-command2"}, deps, false)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Requirement executed and failed"))
	})

	It("returns an error if invalid flag is provided", func() {
		err := NewCommandRunner().Command([]string{"fake-command", "-badFlag"}, deps, false)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Invalid flag: -badFlag"))
	})

})
