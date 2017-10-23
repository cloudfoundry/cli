package pluginaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/actor/pluginaction/pluginactionfakes"
	"code.cloudfoundry.org/cli/api/plugin"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plugin actions", func() {
	var (
		actor            *Actor
		fakeConfig       *pluginactionfakes.FakeConfig
		fakePluginClient *pluginactionfakes.FakePluginClient
	)

	BeforeEach(func() {
		fakeConfig = new(pluginactionfakes.FakeConfig)
		fakePluginClient = new(pluginactionfakes.FakePluginClient)
		actor = NewActor(fakeConfig, fakePluginClient)
	})

	Describe("GetOutdatedPlugins", func() {
		BeforeEach(func() {
			fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{
				{Name: "CF-Community", URL: "https://plugins.cloudfoundry.org"},
				{Name: "Coo Plugins", URL: "https://reallycooplugins.org"},
			})
		})

		Context("when getting a repository errors", func() {
			BeforeEach(func() {
				fakePluginClient.GetPluginRepositoryReturns(plugin.PluginRepository{}, errors.New("generic-error"))
			})

			It("returns a 'GettingPluginRepositoryError", func() {
				_, err := actor.GetOutdatedPlugins()
				Expect(err).To(MatchError(actionerror.GettingPluginRepositoryError{Name: "CF-Community", Message: "generic-error"}))
			})
		})

		Context("when no errors are encountered getting repositories", func() {
			var callNumber int
			BeforeEach(func() {
				callNumber = 0

				fakePluginClient.GetPluginRepositoryStub = func(name string) (plugin.PluginRepository, error) {
					callNumber++
					if callNumber == 1 {
						return plugin.PluginRepository{
							Plugins: []plugin.Plugin{
								{Name: "plugin-1", Version: "2.0.0"},
								{Name: "plugin-2", Version: "1.5.0"},
								{Name: "plugin-3", Version: "3.0.0"},
							},
						}, nil
					} else {
						return plugin.PluginRepository{
							Plugins: []plugin.Plugin{
								{Name: "plugin-1", Version: "1.5.0"},
								{Name: "plugin-2", Version: "2.0.0"},
								{Name: "plugin-3", Version: "3.0.0"},
							},
						}, nil
					}
				}
			})

			Context("when there are outdated plugins", func() {
				BeforeEach(func() {
					fakeConfig.PluginsReturns([]configv3.Plugin{
						{Name: "plugin-1", Version: configv3.PluginVersion{Major: 1, Minor: 0, Build: 0}},
						{Name: "plugin-2", Version: configv3.PluginVersion{Major: 1, Minor: 0, Build: 0}},
						{Name: "plugin-3", Version: configv3.PluginVersion{Major: 3, Minor: 0, Build: 0}},
					})
				})

				It("returns the outdated plugins", func() {
					outdatedPlugins, err := actor.GetOutdatedPlugins()
					Expect(err).ToNot(HaveOccurred())

					Expect(outdatedPlugins).To(Equal([]OutdatedPlugin{
						{Name: "plugin-1", CurrentVersion: "1.0.0", LatestVersion: "2.0.0"},
						{Name: "plugin-2", CurrentVersion: "1.0.0", LatestVersion: "2.0.0"},
					}))
				})
			})

			Context("when there are no outdated plugins", func() {
				BeforeEach(func() {
					fakeConfig.PluginsReturns([]configv3.Plugin{
						{Name: "plugin-1", Version: configv3.PluginVersion{Major: 2, Minor: 0, Build: 0}},
						{Name: "plugin-2", Version: configv3.PluginVersion{Major: 2, Minor: 0, Build: 0}},
						{Name: "plugin-3", Version: configv3.PluginVersion{Major: 3, Minor: 0, Build: 0}},
					})
				})

				It("returns no plugins", func() {
					outdatedPlugins, err := actor.GetOutdatedPlugins()
					Expect(err).ToNot(HaveOccurred())

					Expect(outdatedPlugins).To(BeEmpty())
				})
			})
		})
	})
})
