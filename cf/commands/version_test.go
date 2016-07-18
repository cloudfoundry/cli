package commands_test

import (
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/flags"

	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
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
			cf.Version = "5.0.0"
			cf.Name = "my-special-cf"
			cf.BuiltOnDate = "2016-02-29"
		})

		It("prints the version", func() {
			cmd.Execute(flagContext)

			Expect(ui.Outputs()).To(Equal([]string{
				"my-special-cf version 5.0.0-2016-02-29",
			}))
		})
	})
})
