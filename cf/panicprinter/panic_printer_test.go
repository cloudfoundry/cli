package panicprinter_test

import (
	"code.cloudfoundry.org/cli/cf/errors"
	. "code.cloudfoundry.org/cli/cf/panicprinter"

	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Panic Printer", func() {
	var (
		ui           *terminalfakes.FakeUI
		panicPrinter PanicPrinter
	)

	BeforeEach(func() {
		ui = new(terminalfakes.FakeUI)
		panicPrinter = NewPanicPrinter(ui)
	})

	Describe("DisplayCrashDialog", func() {
		It("includes the error message when given an error", func() {
			panicPrinter.DisplayCrashDialog(errors.New("some-error"), "some command", "some trace")
			Expect(ui.SayCallCount()).To(Equal(1))
			Expect(ui.SayArgsForCall(0)).To(Equal(CrashDialog("some-error", "some command", "some trace")))
		})

		It("includes the string when given a string that is not terminal.QuietPanic", func() {
			panicPrinter.DisplayCrashDialog("some-error", "some command", "some trace")
			Expect(ui.SayCallCount()).To(Equal(1))
			Expect(ui.SayArgsForCall(0)).To(Equal(CrashDialog("some-error", "some command", "some trace")))
		})

		It("prints the unexpected error type message when not given a string or an error", func() {
			panicPrinter.DisplayCrashDialog(struct{}{}, "some command", "some trace")
			Expect(ui.SayCallCount()).To(Equal(1))
			Expect(ui.SayArgsForCall(0)).To(ContainSubstring("An unexpected type of error"))
		})

		It("includes the error message when given an errors.Exception", func() {
			err := errors.Exception{Message: "some-message"}
			panicPrinter.DisplayCrashDialog(err, "some command", "some trace")
			Expect(ui.SayCallCount()).To(Equal(1))
			Expect(ui.SayArgsForCall(0)).To(Equal(CrashDialog(err.Message, "some command", "some trace")))
		})
	})

	Describe("CrashDialog", func() {
		var (
			errMsg      = "the-error-message"
			commandArgs = "command arg1 arg2"
			stackTrace  = "the-stack-trace"
		)

		It("returns crash dialog text", func() {
			Expect(CrashDialog(errMsg, commandArgs, stackTrace)).To(MatchRegexp(`
	Please re-run the command that caused this exception with the environment
	variable CF_TRACE set to true.

	Also, please update to the latest cli and try the command again:
	https://code.cloudfoundry.org/cli/releases

	Please create an issue at: https://code.cloudfoundry.org/cli/issues

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
