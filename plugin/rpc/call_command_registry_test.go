package rpc_test

import (
	"os"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "code.cloudfoundry.org/cli/plugin/rpc"
	. "code.cloudfoundry.org/cli/plugin/rpc/fakecommand"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("calling commands in commandregistry", func() {
	_ = FakeCommand1{} //make sure fake_command is imported and self-registered with init()
	_ = FakeCommand3{} //make sure fake_command is imported and self-registered with init()
	_ = FakeCommand4{} //make sure fake_command is imported and self-registered with init()

	var (
		ui         *terminalfakes.FakeUI
		deps       commandregistry.Dependency
		fakeLogger *tracefakes.FakePrinter
	)

	BeforeEach(func() {
		fakeLogger = new(tracefakes.FakePrinter)
		deps = commandregistry.NewDependency(os.Stdout, fakeLogger, "")
		ui = new(terminalfakes.FakeUI)
		deps.UI = ui

		cmd := commandregistry.Commands.FindCommand("fake-command")
		commandregistry.Commands.SetCommand(cmd.SetDependency(deps, true))

		cmd2 := commandregistry.Commands.FindCommand("fake-command2")
		commandregistry.Commands.SetCommand(cmd2.SetDependency(deps, true))
	})

	Context("when command exists and the correct flags are passed", func() {
		BeforeEach(func() {
			err := NewCommandRunner().Command([]string{"fake-command"}, deps, false)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should set dependencies, execute requirements, and execute the command", func() {
			Expect(ui.SayArgsForCall(0)).To(ContainSubstring("SetDependency() called, pluginCall true"))
			Expect(ui.SayArgsForCall(1)).To(ContainSubstring("SetDependency() called, pluginCall false"))
			Expect(ui.SayArgsForCall(2)).To(ContainSubstring("Requirement executed"))
			Expect(ui.SayArgsForCall(3)).To(ContainSubstring("Command Executed"))
		})
	})

	Context("when any of the command requirements fails", func() {
		It("returns an error", func() {
			err := NewCommandRunner().Command([]string{"fake-command2"}, deps, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Requirement executed and failed"))
		})
	})

	Context("when invalid flags are provided", func() {
		It("returns an error", func() {
			err := NewCommandRunner().Command([]string{"fake-command", "-badFlag"}, deps, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid flag: -badFlag"))
		})
	})

	Context("when the command execute errors", func() {
		BeforeEach(func() {
			cmd4 := commandregistry.Commands.FindCommand("fake-command4")
			commandregistry.Commands.SetCommand(cmd4.SetDependency(deps, true))
		})

		It("returns an error", func() {
			err := NewCommandRunner().Command([]string{"fake-command4"}, deps, false)
			Expect(err).To(MatchError(ErrFakeCommand4))
		})
	})

	Context("when the command execute panics", func() {
		BeforeEach(func() {
			cmd3 := commandregistry.Commands.FindCommand("fake-command3")
			commandregistry.Commands.SetCommand(cmd3.SetDependency(deps, true))
		})

		It("returns an error", func() {
			err := NewCommandRunner().Command([]string{"fake-command3"}, deps, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("cli_rpc_server_test"))
		})
	})
})
