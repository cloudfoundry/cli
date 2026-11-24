package plugin_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/actor/pluginaction"
	"code.cloudfoundry.org/cli/v8/api/plugin"
	"code.cloudfoundry.org/cli/v8/command/commandfakes"
	. "code.cloudfoundry.org/cli/v8/command/plugin"
	"code.cloudfoundry.org/cli/v8/command/plugin/pluginfakes"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("repo-plugins command", func() {
	var (
		cmd        RepoPluginsCommand
		testUI     *ui.UI
		fakeConfig *commandfakes.FakeConfig
		fakeActor  *pluginfakes.FakeRepoPluginsActor
		executeErr error
		repos      []configv3.PluginRepository
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(pluginfakes.FakeRepoPluginsActor)
		cmd = RepoPluginsCommand{UI: testUI, Config: fakeConfig, Actor: fakeActor}

		repos = []configv3.PluginRepository{
			{Name: "repo1", URL: "https://repo1.example.com"},
			{Name: "repo2", URL: "https://repo2.example.com"},
		}
		fakeConfig.PluginRepositoriesReturns(repos)

		fakeActor.GetPluginsFromRepositoriesReturns([]pluginaction.RepositoryPlugins{
			{RepositoryName: "repo1", Plugins: []plugin.Plugin{{Name: "plugin1", Version: "1.0.0", Description: "desc1"}}},
		}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("no repository is specified", func() {
		It("lists the plugins of all registered repositories", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.GetPluginsFromRepositoriesCallCount()).To(Equal(1))
			Expect(fakeActor.GetPluginsFromRepositoriesArgsForCall(0)).To(Equal(repos))

			Expect(testUI.Out).To(Say(`Getting plugins from all repositories \.\.\.`))
			Expect(testUI.Out).To(Say("Repository: repo1"))
			Expect(testUI.Out).To(Say(`name\s+version\s+description`))
			Expect(testUI.Out).To(Say(`plugin1\s+1\.0\.0\s+desc1`))
		})
	})

	When("a registered repository is specified with -r", func() {
		BeforeEach(func() {
			cmd.RegisteredRepository = "repo2"
			fakeActor.GetPluginRepositoryReturns(repos[1], nil)
		})

		It("lists only that repository's plugins", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.GetPluginRepositoryArgsForCall(0)).To(Equal("repo2"))
			Expect(fakeActor.GetPluginsFromRepositoriesArgsForCall(0)).To(Equal([]configv3.PluginRepository{repos[1]}))
			Expect(testUI.Out).To(Say(`Getting plugins from repository 'repo2' \.\.\.`))
		})
	})

	When("an unregistered repository is specified with -r", func() {
		BeforeEach(func() {
			cmd.RegisteredRepository = "unknown"
			fakeActor.GetPluginRepositoryReturns(configv3.PluginRepository{}, actionerror.RepositoryNotRegisteredError{Name: "unknown"})
		})

		It("returns the error without fetching any plugins", func() {
			Expect(executeErr).To(MatchError(actionerror.RepositoryNotRegisteredError{Name: "unknown"}))
			Expect(fakeActor.GetPluginsFromRepositoriesCallCount()).To(Equal(0))
		})
	})

	When("some repositories cannot be fetched", func() {
		BeforeEach(func() {
			fakeActor.GetPluginsFromRepositoriesReturns(nil, []error{
				actionerror.FetchingPluginInfoFromRepositoryError{RepositoryName: "repo2", Err: errors.New("404 not found")},
			})
		})

		It("logs the errors on stderr and still succeeds", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Err).To(Say("Logged errors:"))
			Expect(testUI.Err).To(Say("Plugin repository repo2 returned 404 not found."))
		})
	})
})
