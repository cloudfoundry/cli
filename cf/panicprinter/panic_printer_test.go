package panicprinter_test

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/panicprinter"
	"github.com/cloudfoundry/cli/cf/terminal"

	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Panic Printer", func() {
	var ui *testterm.FakeUI

	BeforeEach(func() {
		panicprinter.UI = &testterm.FakeUI{}
		ui = panicprinter.UI.(*testterm.FakeUI)
	})

	Describe("DisplayCrashDialog", func() {
		It("includes the error message when given an error", func() {
			panicprinter.DisplayCrashDialog(errors.New("some-error"), "some command", "some trace")
			Expect(len(ui.Outputs())).To(BeNumerically(">", 0))
			Expect(ui.Outputs()).To(ContainElement(ContainSubstring("some-error")))
		})

		It("includes the string when given a string that is not terminal.QuietPanic", func() {
			panicprinter.DisplayCrashDialog("some-error", "some command", "some trace")
			Expect(len(ui.Outputs())).To(BeNumerically(">", 0))
			Expect(ui.Outputs()).To(ContainElement(ContainSubstring("some-error")))
		})

		It("does not print anything when given a string that is terminal.QuietPanic", func() {
			err := terminal.QuietPanic
			panicprinter.DisplayCrashDialog(err, "some command", "some trace")
			Expect(len(ui.Outputs())).To(Equal(0))
		})

		It("prints the unexpected error type message when not given a string or an error", func() {
			panicprinter.DisplayCrashDialog(struct{}{}, "some command", "some trace")
			Expect(len(ui.Outputs())).To(BeNumerically(">", 0))
			Expect(ui.Outputs()).To(ContainElement(ContainSubstring("An unexpected type of error")))
		})

		It("includes the error message when given an errors.Exception", func() {
			err := errors.Exception{Message: "some-message"}
			panicprinter.DisplayCrashDialog(err, "some command", "some trace")
			Expect(len(ui.Outputs())).To(BeNumerically(">", 0))
			Expect(ui.Outputs()).To(ContainElement(ContainSubstring("some-message")))
		})
	})

	Describe("CrashDialog", func() {
		var (
			errMsg      = "the-error-message"
			commandArgs = "command arg1 arg2"
			stackTrace  = "the-stack-trace"
		)

		It("returns crash dialog text", func() {
			Expect(panicprinter.CrashDialog(errMsg, commandArgs, stackTrace)).To(MatchRegexp(`
	Please re-run the command that caused this exception with the environment
	variable CF_TRACE set to true.

	Also, please update to the latest cli and try the command again:
	https://github.com/cloudfoundry/cli/releases

	Please create an issue at: https://github.com/cloudfoundry/cli/issues

	Include the below information when creating the issue:

		Command
		command arg1 arg2

		CLI Version
		.*

		Error
		the-error-message

		Stack Trace
		the-stack-trace

		Your Platform Details
		e.g. Mac OS X 10.11, Windows 8.1 64-bit, Ubuntu 14.04.3 64-bit

		Shell
		e.g. Terminal, iTerm, Powershell, Cygwin, gnome-terminal, terminator
`))
		})
	})
})
