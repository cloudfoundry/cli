package plugin_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/plugin"
	"code.cloudfoundry.org/cli/command/plugin/pluginfakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("add-plugin-repo command", func() {
	var (
		cmd        AddPluginRepoCommand
		testUI     *ui.UI
		fakeConfig *commandfakes.FakeConfig
		fakeActor  *pluginfakes.FakeAddPluginRepoActor
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(pluginfakes.FakeAddPluginRepoActor)
		cmd = AddPluginRepoCommand{UI: testUI, Config: fakeConfig, Actor: fakeActor}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the provided repo name already exists", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.PluginRepoName = "some-repo"
			cmd.RequiredArgs.PluginRepoURL = "some-repo-URL"
			fakeActor.AddPluginRepositoryReturns(actionerror.RepositoryNameTakenError{Name: "some-repo"})
		})

		It("returns RepositoryNameTakenError", func() {
			Expect(executeErr).To(MatchError(actionerror.RepositoryNameTakenError{Name: "some-repo"}))
		})
	})

	Context("when the provided repo name and URL already exist in the same repo", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.PluginRepoName = "some-repo"
			cmd.RequiredArgs.PluginRepoURL = "some-repo-URL"
			fakeActor.AddPluginRepositoryReturns(actionerror.RepositoryAlreadyExistsError{Name: "some-repo", URL: "https://some-repo-URL"})
		})

		It("displays a message that the repo is already registered and does not return an error", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("https://some-repo-URL already registered as some-repo"))

			Expect(fakeActor.AddPluginRepositoryCallCount()).To(Equal(1))
			repoName, repoURL := fakeActor.AddPluginRepositoryArgsForCall(0)
			Expect(repoName).To(Equal("some-repo"))
			Expect(repoURL).To(Equal("some-repo-URL"))
		})
	})

	Context("when an AddPluginRepositoryError is encountered", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.PluginRepoName = "some-repo"
			cmd.RequiredArgs.PluginRepoURL = "some-URL"
			fakeActor.AddPluginRepositoryReturns(actionerror.AddPluginRepositoryError{Name: "some-repo", URL: "some-URL", Message: "404"})
		})

		It("handles the error", func() {
			Expect(executeErr).To(MatchError(actionerror.AddPluginRepositoryError{Name: "some-repo", URL: "some-URL", Message: "404"}))
		})
	})

	Context("when no errors are encountered", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.PluginRepoName = "some-repo"
			cmd.RequiredArgs.PluginRepoURL = "https://some-repo-URL"
			fakeActor.AddPluginRepositoryReturns(nil)
		})

		It("adds the plugin repo", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("https://some-repo-URL added as some-repo"))

			Expect(fakeActor.AddPluginRepositoryCallCount()).To(Equal(1))
			repoName, repoURL := fakeActor.AddPluginRepositoryArgsForCall(0)
			Expect(repoName).To(Equal("some-repo"))
			Expect(repoURL).To(Equal("https://some-repo-URL"))
		})
	})
})
