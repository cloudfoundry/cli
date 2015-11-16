package panic_printer_test

import (
	"github.com/cloudfoundry/cli/cf"
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
		var errMsg = "this is an error"
		var commandArgs = "command line arguments"
		var stackTrace = "1000 bottles of beer"

		It("should return a string containing the default error text", func() {
			Expect(panic_printer.CrashDialog(errMsg, commandArgs, stackTrace)).To(ContainSubstring("Please file this bug : https://github.com/cloudfoundry/cli/issues"))
		})

		It("should return the command name", func() {
			Expect(panic_printer.CrashDialog(errMsg, commandArgs, stackTrace)).To(ContainSubstring(cf.Name()))
		})

		It("should return the inputted arguments", func() {
			Expect(panic_printer.CrashDialog(errMsg, commandArgs, stackTrace)).To(ContainSubstring("command line arguments"))
		})

		It("should return the specific error message", func() {
			Expect(panic_printer.CrashDialog(errMsg, commandArgs, stackTrace)).To(ContainSubstring("this is an error"))
		})

		It("should return the stack trace", func() {
			Expect(panic_printer.CrashDialog(errMsg, commandArgs, stackTrace)).To(ContainSubstring("1000 bottles of beer"))
		})

		It("should print the cli version", func() {
			Expect(panic_printer.CrashDialog(errMsg, commandArgs, stackTrace)).To(ContainSubstring(cf.Version))
		})
	})
})
