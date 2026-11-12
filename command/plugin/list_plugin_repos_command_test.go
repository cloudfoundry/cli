package plugin_test

import (
	"code.cloudfoundry.org/cli/v8/command/commandfakes"
	. "code.cloudfoundry.org/cli/v8/command/plugin"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("list-plugin-repos command", func() {
	var (
		cmd        ListPluginReposCommand
		testUI     *ui.UI
		fakeConfig *commandfakes.FakeConfig
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		cmd = ListPluginReposCommand{UI: testUI, Config: fakeConfig}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("plugin repositories are registered", func() {
		BeforeEach(func() {
			fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{
				{Name: "repo-1", URL: "https://repo1.example.com"},
				{Name: "repo-2", URL: "https://repo2.example.com"},
			})
		})

		It("displays OK and a table of the repositories", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("OK"))
			Expect(testUI.Out).To(Say(`Repo Name\s+URL`))
			Expect(testUI.Out).To(Say(`repo-1\s+https://repo1.example.com`))
			Expect(testUI.Out).To(Say(`repo-2\s+https://repo2.example.com`))
		})
	})

	When("no plugin repositories are registered", func() {
		It("displays OK and only the table header", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("OK"))
			Expect(testUI.Out).To(Say(`Repo Name\s+URL`))
		})
	})
})
