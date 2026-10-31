package plugin_test

import (
	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/command/commandfakes"
	. "code.cloudfoundry.org/cli/v8/command/plugin"
	"code.cloudfoundry.org/cli/v8/command/plugin/pluginfakes"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("remove-plugin-repo command", func() {
	var (
		cmd        RemovePluginRepoCommand
		testUI     *ui.UI
		fakeConfig *commandfakes.FakeConfig
		fakeActor  *pluginfakes.FakeRemovePluginRepoActor
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(pluginfakes.FakeRemovePluginRepoActor)
		cmd = RemovePluginRepoCommand{UI: testUI, Config: fakeConfig, Actor: fakeActor}
		cmd.RequiredArgs.PluginRepoName = "some-repo"
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the repository is registered", func() {
		It("removes it and reports success", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.RemovePluginRepositoryCallCount()).To(Equal(1))
			Expect(fakeActor.RemovePluginRepositoryArgsForCall(0)).To(Equal("some-repo"))

			Expect(testUI.Out).To(Say("OK"))
			Expect(testUI.Out).To(Say("some-repo removed from list of repositories"))
		})
	})

	When("the repository is not registered", func() {
		BeforeEach(func() {
			fakeActor.RemovePluginRepositoryReturns(actionerror.RepositoryNotRegisteredError{Name: "some-repo"})
		})

		It("returns the error and does not report success", func() {
			Expect(executeErr).To(MatchError(actionerror.RepositoryNotRegisteredError{Name: "some-repo"}))
			Expect(testUI.Out).ToNot(Say("removed from list of repositories"))
		})
	})
})
