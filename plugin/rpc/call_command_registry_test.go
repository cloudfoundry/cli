package rpc_test

import (
	"github.com/cloudfoundry/cli/cf/commandregistry"
	. "github.com/cloudfoundry/cli/plugin/rpc"
	. "github.com/cloudfoundry/cli/plugin/rpc/fakecommand"

	"github.com/cloudfoundry/cli/cf/trace/tracefakes"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("calling commands in commandregistry", func() {

	_ = FakeCommand1{} //make sure fake_command is imported and self-registered with init()

	var (
		ui         *testterm.FakeUI
		deps       commandregistry.Dependency
		fakeLogger *tracefakes.FakePrinter
	)

	BeforeEach(func() {
		fakeLogger = new(tracefakes.FakePrinter)
		deps = commandregistry.NewDependency(os.Stdout, fakeLogger)
		ui = &testterm.FakeUI{}
		deps.UI = ui

		cmd := commandregistry.Commands.FindCommand("fake-command")
		commandregistry.Commands.SetCommand(cmd.SetDependency(deps, true))

		cmd2 := commandregistry.Commands.FindCommand("fake-command2")
		commandregistry.Commands.SetCommand(cmd2.SetDependency(deps, true))
	})

	It("runs the command requirements", func() {
		NewCommandRunner().Command([]string{"fake-command"}, deps, false)
		Expect(ui.Outputs()).To(ContainSubstrings([]string{"Requirement executed"}))
	})

	It("calls the command Execute() func", func() {
		NewCommandRunner().Command([]string{"fake-command"}, deps, false)
		Expect(ui.Outputs()).To(ContainSubstrings([]string{"Command Executed"}))
	})

	It("sets the dependency of the command", func() {
		NewCommandRunner().Command([]string{"fake-command"}, deps, false)
		Expect(ui.Outputs()).To(ContainSubstrings([]string{"SetDependency() called, pluginCall true"}))
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
