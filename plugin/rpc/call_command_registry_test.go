package rpc_test

import (
	"os"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/terminal/terminalfakes"
	"github.com/cloudfoundry/cli/cf/trace/tracefakes"
	. "github.com/cloudfoundry/cli/plugin/rpc"
	. "github.com/cloudfoundry/cli/plugin/rpc/fakecommand"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("calling commands in commandregistry", func() {

	_ = FakeCommand1{} //make sure fake_command is imported and self-registered with init()

	var (
		ui         *terminalfakes.FakeUI
		deps       commandregistry.Dependency
		fakeLogger *tracefakes.FakePrinter
	)

	BeforeEach(func() {
		fakeLogger = new(tracefakes.FakePrinter)
		deps = commandregistry.NewDependency(os.Stdout, fakeLogger)
		ui = new(terminalfakes.FakeUI)
		deps.UI = ui

		cmd := commandregistry.Commands.FindCommand("fake-command")
		commandregistry.Commands.SetCommand(cmd.SetDependency(deps, true))

		cmd2 := commandregistry.Commands.FindCommand("fake-command2")
		commandregistry.Commands.SetCommand(cmd2.SetDependency(deps, true))
	})

	Context("when not expecting the command to fail", func() {
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

	Context("when expecting the command to fail", func() {
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
})
