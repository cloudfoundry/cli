package panic_printer_test

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/panic_printer"
	"github.com/cloudfoundry/cli/cf/terminal"

	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Panic Printer", func() {
	var ui *testterm.FakeUI

	BeforeEach(func() {
		panic_printer.UI = &testterm.FakeUI{}
		ui = panic_printer.UI.(*testterm.FakeUI)
	})

	Describe("DisplayCrashDialog", func() {
		It("includes the error message when given an error", func() {
			panic_printer.DisplayCrashDialog(errors.New("some-error"), "some command", "some trace")
			Expect(len(ui.Outputs)).To(BeNumerically(">", 0))
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("some-error")))
		})

		It("includes the string when given a string that is not terminal.QuietPanic", func() {
			panic_printer.DisplayCrashDialog("some-error", "some command", "some trace")
			Expect(len(ui.Outputs)).To(BeNumerically(">", 0))
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("some-error")))
		})

		It("does not print anything when given a string that is terminal.QuietPanic", func() {
			err := terminal.QuietPanic
			panic_printer.DisplayCrashDialog(err, "some command", "some trace")
			Expect(len(ui.Outputs)).To(Equal(0))
		})

		It("prints the unexpected error type message when not given a string or an error", func() {
			panic_printer.DisplayCrashDialog(struct{}{}, "some command", "some trace")
			Expect(len(ui.Outputs)).To(BeNumerically(">", 0))
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("An unexpected type of error")))
		})

		It("includes the error message when given an errors.Exception with DisplayCrashDialog set to true", func() {
			err := errors.Exception{DisplayCrashDialog: true, Message: "some-message"}
			panic_printer.DisplayCrashDialog(err, "some command", "some trace")
			Expect(len(ui.Outputs)).To(BeNumerically(">", 0))
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("some-message")))
		})

		It("does not print anything when given an errors.Exception with DisplayCrashDialog set to false", func() {
			err := errors.Exception{DisplayCrashDialog: false, Message: "some-message"}
			panic_printer.DisplayCrashDialog(err, "some command", "some trace")
			Expect(len(ui.Outputs)).To(Equal(0))
		})
	})

	Describe("CrashDialog", func() {
		var (
			errMsg      = "the-error-message"
			commandArgs = "command arg1 arg2"
			stackTrace  = "the-stack-trace"
		)

		It("returns crash dialog text", func() {
			Expect(panic_printer.CrashDialog(errMsg, commandArgs, stackTrace)).To(MatchRegexp(`
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
