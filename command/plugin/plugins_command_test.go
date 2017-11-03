package plugin_test

import (
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/plugin"
	"code.cloudfoundry.org/cli/command/plugin/pluginfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("plugins Command", func() {
	var (
		cmd        PluginsCommand
		testUI     *ui.UI
		fakeConfig *commandfakes.FakeConfig
		executeErr error
		fakeActor  *pluginfakes.FakePluginsActor
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(pluginfakes.FakePluginsActor)
		cmd = PluginsCommand{UI: testUI, Config: fakeConfig, Actor: fakeActor}
		cmd.Checksum = false

		fakeConfig.BinaryNameReturns("faceman")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when there are no plugins installed", func() {
		It("displays the empty table", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(testUI.Out).To(Say("Listing installed plugins..."))
			Expect(testUI.Out).To(Say(""))
			Expect(testUI.Out).To(Say("plugin\\s+version\\s+command name\\s+command help"))
			Expect(testUI.Out).To(Say(""))
			Expect(testUI.Out).To(Say("Use 'faceman repo-plugins' to list plugins in registered repos available to install\\."))
			Expect(testUI.Out).ToNot(Say("[A-Za-z0-9]+"))
		})

		Context("when the --checksum flag is provided", func() {
			BeforeEach(func() {
				cmd.Checksum = true
			})

			It("displays the empty checksums table", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say("Computing sha1 for installed plugins, this may take a while..."))
				Expect(testUI.Out).To(Say(""))
				Expect(testUI.Out).To(Say("plugin\\s+version\\s+sha1"))
				Expect(testUI.Out).ToNot(Say("[A-Za-z0-9]+"))
			})
		})

	})

	Context("when there are plugins installed", func() {
		var plugins []configv3.Plugin

		BeforeEach(func() {
			plugins = []configv3.Plugin{
				{
					Name: "Sorted-first",
					Version: configv3.PluginVersion{
						Major: 1,
						Minor: 1,
						Build: 0,
					},
					Commands: []configv3.PluginCommand{
						{
							Name:     "command-2",
							HelpText: "help-command-2",
						},
						{
							Name:     "command-1",
							Alias:    "c",
							HelpText: "help-command-1",
						},
					},
				},
				{
					Name: "sorted-second",
					Version: configv3.PluginVersion{
						Major: 0,
						Minor: 0,
						Build: 0,
					},
					Commands: []configv3.PluginCommand{
						{
							Name:     "foo",
							HelpText: "help-foo",
						},
						{
							Name:     "bar",
							HelpText: "help-bar",
						},
					},
				},
			}
			fakeConfig.PluginsReturns(plugins)
		})

		It("displays the plugins in alphabetical order and their commands", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(testUI.Out).To(Say("Listing installed plugins..."))
			Expect(testUI.Out).To(Say(""))
			Expect(testUI.Out).To(Say("plugin\\s+version\\s+command name\\s+command help"))
			Expect(testUI.Out).To(Say("Sorted-first\\s+1\\.1\\.0\\s+command-1, c\\s+help-command-1"))
			Expect(testUI.Out).To(Say("Sorted-first\\s+1\\.1\\.0\\s+command-2\\s+help-command-2"))
			Expect(testUI.Out).To(Say("sorted-second\\s+N/A\\s+bar\\s+help-bar"))
			Expect(testUI.Out).To(Say("sorted-second\\s+N/A\\s+foo\\s+help-foo"))
			Expect(testUI.Out).To(Say(""))
			Expect(testUI.Out).To(Say("Use 'faceman repo-plugins' to list plugins in registered repos available to install\\."))
		})

		Context("when the --checksum flag is provided", func() {
			var (
				file *os.File
			)

			BeforeEach(func() {
				cmd.Checksum = true

				var err error
				file, err = ioutil.TempFile("", "")
				defer file.Close()
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(file.Name(), []byte("some-text"), 0600)
				Expect(err).NotTo(HaveOccurred())

				plugins[0].Location = file.Name()

				plugins[1].Location = "/wut/wut/"
			})

			AfterEach(func() {
				err := os.Remove(file.Name())
				Expect(err).NotTo(HaveOccurred())
			})

			It("displays the plugin checksums", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say("Computing sha1 for installed plugins, this may take a while..."))
				Expect(testUI.Out).To(Say(""))
				Expect(testUI.Out).To(Say("plugin\\s+version\\s+sha1"))
				Expect(testUI.Out).To(Say("Sorted-first\\s+1\\.1\\.0\\s+2142a57cb8587400fa7f4ee492f25cf07567f4a5"))
				Expect(testUI.Out).To(Say("sorted-second\\s+N/A\\s+N/A"))
			})
		})

		Context("when the --outdated flag is provided", func() {
			BeforeEach(func() {
				cmd.Outdated = true
			})

			Context("when there are no repositories", func() {
				BeforeEach(func() {
					fakeConfig.PluginRepositoriesReturns(nil)
				})

				It("returns the 'No plugin repositories added' error", func() {
					Expect(executeErr).To(MatchError(translatableerror.NoPluginRepositoriesError{}))
					Expect(testUI.Out).NotTo(Say("Searching"))
				})
			})

			Context("when there are repositories", func() {
				BeforeEach(func() {
					fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{
						{Name: "repo-1", URL: "https://repo-1.plugins.com"},
						{Name: "repo-2", URL: "https://repo-2.plugins.com"},
					})
				})

				Context("when the actor returns GettingRepositoryError", func() {
					BeforeEach(func() {
						fakeActor.GetOutdatedPluginsReturns(nil, actionerror.GettingPluginRepositoryError{
							Name:    "repo-1",
							Message: "404",
						})
					})
					It("displays the repository and the error", func() {
						Expect(executeErr).To(MatchError(actionerror.GettingPluginRepositoryError{
							Name:    "repo-1",
							Message: "404",
						}))

						Expect(testUI.Out).To(Say("Searching repo-1, repo-2 for newer versions of installed plugins..."))
					})
				})

				Context("when there are no outdated plugins", func() {
					It("displays the empty outdated table", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(testUI.Out).To(Say("Searching repo-1, repo-2 for newer versions of installed plugins..."))
						Expect(testUI.Out).To(Say(""))
						Expect(testUI.Out).To(Say("plugin\\s+version\\s+latest version\\n\\nUse 'faceman install-plugin' to update a plugin to the latest version\\."))

						Expect(fakeActor.GetOutdatedPluginsCallCount()).To(Equal(1))
					})
				})

				Context("when plugins are outdated", func() {
					BeforeEach(func() {
						fakeActor.GetOutdatedPluginsReturns([]pluginaction.OutdatedPlugin{
							{Name: "plugin-1", CurrentVersion: "1.0.0", LatestVersion: "2.0.0"},
							{Name: "plugin-2", CurrentVersion: "2.0.0", LatestVersion: "3.0.0"},
						}, nil)
					})

					It("displays the outdated plugins", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(fakeActor.GetOutdatedPluginsCallCount()).To(Equal(1))

						Expect(testUI.Out).To(Say("Searching repo-1, repo-2 for newer versions of installed plugins..."))
						Expect(testUI.Out).To(Say(""))
						Expect(testUI.Out).To(Say("plugin\\s+version\\s+latest version"))
						Expect(testUI.Out).To(Say("plugin-1\\s+1.0.0\\s+2.0.0"))
						Expect(testUI.Out).To(Say("plugin-2\\s+2.0.0\\s+3.0.0"))
						Expect(testUI.Out).To(Say(""))
						Expect(testUI.Out).To(Say("Use 'faceman install-plugin' to update a plugin to the latest version\\."))
					})
				})
			})
		})
	})
})
