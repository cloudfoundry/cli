package v2_test

import (
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
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
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		cmd = PluginsCommand{UI: testUI, Config: fakeConfig}
		cmd.Checksum = false
		fakeConfig.ExperimentalReturns(true)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when there are no plugins installed", func() {
		It("displays the empty table", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Eventually(testUI.Out).Should(Say("Listing installed plugins..."))
			Eventually(testUI.Out).Should(Say(""))
			Eventually(testUI.Out).Should(Say("plugin name\\s+version\\s+command name\\s+command help"))
			Consistently(testUI.Out).ShouldNot(Say("[A-Za-z0-9]+"))
		})

		Context("when the --checksum flag is provided", func() {
			BeforeEach(func() {
				cmd.Checksum = true
			})

			It("displays the empty checksums table", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Eventually(testUI.Out).Should(Say("Computing sha1 for installed plugins, this may take a while..."))
				Eventually(testUI.Out).Should(Say(""))
				Eventually(testUI.Out).Should(Say("plugin name\\s+version\\s+sha1"))
				Consistently(testUI.Out).ShouldNot(Say("[A-Za-z0-9]+"))
			})
		})
	})

	Context("when there are plugins installed", func() {
		Context("when there are multiple plugins", func() {
			BeforeEach(func() {
				fakeConfig.PluginsReturns(map[string]configv3.Plugin{
					"sorted-second": {
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
					"Sorted-first": {
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
				})
			})

			It("displays the plugins in alphabetical order and their commands", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Eventually(testUI.Out).Should(Say("Listing installed plugins..."))
				Eventually(testUI.Out).Should(Say(""))
				Eventually(testUI.Out).Should(Say("plugin name\\s+version\\s+command name\\s+command help"))
				Eventually(testUI.Out).Should(Say("Sorted-first\\s+1\\.1\\.0\\s+command-2\\s+help-command-2"))
				Eventually(testUI.Out).Should(Say("Sorted-first\\s+1\\.1\\.0\\s+command-1, c\\s+help-command-1"))
				Eventually(testUI.Out).Should(Say("sorted-second\\s+N/A\\s+foo\\s+help-foo"))
				Eventually(testUI.Out).Should(Say("sorted-second\\s+N/A\\s+bar\\s+help-bar"))
			})

			XContext("when the --checksum flag is provided", func() {
				BeforeEach(func() {
					cmd.Checksum = true
				})

				It("displays the plugin checksums", func() {
				})

				Context("when an error is encountered calculating the sha1 of a plugin", func() {
					It("displays N/A for the sha1 of that plugin", func() {
					})
				})
			})
		})
	})
})
