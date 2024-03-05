package commands_test

import (
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands"
	"code.cloudfoundry.org/cli/cf/flags"

	testterm "code.cloudfoundry.org/cli/cf/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("version Command", func() {
	var (
		ui  *testterm.FakeUI
		cmd commandregistry.Command
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		deps := commandregistry.Dependency{
			UI: ui,
		}

		cmd = &commands.Version{}
		cmd.SetDependency(deps, false)
	})

	Describe("Execute", func() {
		var flagContext flags.FlagContext

		BeforeEach(func() {
			cf.Name = "my-special-cf"
		})

		It("prints the version", func() {
			cmd.Execute(flagContext)

			Expect(ui.Outputs()).To(Equal([]string{
				"my-special-cf version 0.0.0-unknown-version",
			}))
		})
	})
})
