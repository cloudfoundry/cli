package panic_printer_test

import (
	"github.com/cloudfoundry/cli/cf"
	. "github.com/cloudfoundry/cli/cf/panic_printer"
	"github.com/cloudfoundry/cli/cf/terminal"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Panic Printer", func() {
	var ui *testterm.FakeUI

	BeforeEach(func() {
		UI = &testterm.FakeUI{}
		ui = UI.(*testterm.FakeUI)
	})

	Describe("DisplayCrashDialog", func() {
		Context("when given an err set to QuietPanic", func() {
			It("should not print anything", func() {
				err := terminal.QuietPanic
				DisplayCrashDialog(err, "some command", "some trace")
				Expect(len(ui.Outputs)).To(Equal(0))
			})
		})
	})

	Describe("CrashDialog", func() {
		var errMsg = "this is an error"
		var commandArgs = "command line arguments"
		var stackTrace = "1000 bottles of beer"

		It("should return a string containing the default error text", func() {
			Expect(CrashDialog(errMsg, commandArgs, stackTrace)).To(ContainSubstring("Please file this bug : https://github.com/cloudfoundry/cli/issues"))
		})

		It("should return the command name", func() {
			Expect(CrashDialog(errMsg, commandArgs, stackTrace)).To(ContainSubstring(cf.Name()))
		})

		It("should return the inputted arguments", func() {
			Expect(CrashDialog(errMsg, commandArgs, stackTrace)).To(ContainSubstring("command line arguments"))
		})

		It("should return the specific error message", func() {
			Expect(CrashDialog(errMsg, commandArgs, stackTrace)).To(ContainSubstring("this is an error"))
		})

		It("should return the stack trace", func() {
			Expect(CrashDialog(errMsg, commandArgs, stackTrace)).To(ContainSubstring("1000 bottles of beer"))
		})

		It("should print the cli version", func() {
			Expect(CrashDialog(errMsg, commandArgs, stackTrace)).To(ContainSubstring(cf.Version))
		})
	})
})
