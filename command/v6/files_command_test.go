package v6_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Files Command", func() {
	var (
		cmd    FilesCommand
		testUI *ui.UI
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())

		cmd = FilesCommand{
			UI: testUI,
		}
	})

	When("command is called in v6", func() {
		It("returns an unrefactored command error", func() {
			err := cmd.Execute(nil)
			Expect(err).To(MatchError(translatableerror.UnrefactoredCommandError{}))
		})

		It("displays a deprecation warning", func() {
			_ = cmd.Execute(nil)
			Expect(testUI.Err).To(Say("Deprecation warning: This command has been deprecated and will be removed in the future. For similar functionality, please use the `cf ssh` command instead."))
		})
	})
})
