package v2_test

import (
	"io/ioutil"
	"os"

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
		var plugins map[string]configv3.Plugin

		BeforeEach(func() {
			plugins = map[string]configv3.Plugin{
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
			}
			fakeConfig.PluginsReturns(plugins)
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

		Context("when the --checksum flag is provided", func() {
			var (
				file    *os.File
				dirPath string
			)

			BeforeEach(func() {
				cmd.Checksum = true

				var err error
				file, err = ioutil.TempFile("", "")
				defer file.Close()
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(file.Name(), []byte("some-text"), 0600)
				Expect(err).NotTo(HaveOccurred())

				plugin1 := plugins["Sorted-first"]
				plugin1.Location = file.Name()
				plugins["Sorted-first"] = plugin1

				dirPath, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				plugin2 := plugins["sorted-second"]
				plugin2.Location = dirPath
				plugins["sorted-second"] = plugin2
			})

			AfterEach(func() {
				err := os.Remove(file.Name())
				Expect(err).NotTo(HaveOccurred())
				err = os.RemoveAll(dirPath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("displays the plugin checksums", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Eventually(testUI.Out).Should(Say("Computing sha1 for installed plugins, this may take a while..."))
				Eventually(testUI.Out).Should(Say(""))
				Eventually(testUI.Out).Should(Say("plugin name\\s+version\\s+sha1"))
				Eventually(testUI.Out).Should(Say("Sorted-first\\s+1\\.1\\.0\\s+2142a57cb8587400fa7f4ee492f25cf07567f4a5"))
				Eventually(testUI.Out).Should(Say("sorted-second\\s+N/A\\s+N/A"))
			})
		})
	})
})
