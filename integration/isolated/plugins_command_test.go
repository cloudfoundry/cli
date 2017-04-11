package isolated

import (
	"fmt"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("plugins command", func() {
	Describe("help", func() {
		Context("when --help flag is provided", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("plugins", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("plugins - List all available plugin commands"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf plugins [--checksum]"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--checksum\\s+Compute and show the sha1 value of the plugin binary file"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("install-plugin, repo-plugins, uninstall-plugin"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when no plugins are installed", func() {
		It("displays an empty table", func() {
			session := helpers.CF("plugins")
			Eventually(session).Should(Say("plugin name\\s+version\\s+command name\\s+command help"))
			Consistently(session).ShouldNot(Say("[a-za-z0-9]+"))
			Eventually(session).Should(Exit(0))
		})

		Context("when the --checksum flag is provided", func() {
			It("displays an empty checksum table", func() {
				session := helpers.CF("plugins", "--checksum")
				Eventually(session).Should(Say("plugin name\\s+sha1"))
				Consistently(session).ShouldNot(Say("[a-za-z0-9]+"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when plugins are installed", func() {
		var (
			pathToPlugins string
			pluginConfig  string
		)

		BeforeEach(func() {
			pathToPlugins = fmt.Sprintf("%s/.cf/plugins", homeDir)
			err := os.MkdirAll(pathToPlugins, 0700)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(pathToPlugins)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			err := ioutil.WriteFile(fmt.Sprintf("%s/config.json", pathToPlugins), []byte(pluginConfig), 0600)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there are multiple plugins", func() {
			BeforeEach(func() {
				pluginConfig = `{
					"Plugins": {
						"sorted-third": {
							"Location": "plugin-dir/plugin-3",
							"Version": {
								"Major": 2,
								"Minor": 0,
								"Build": 1
							},
							"Commands": [
								{
									"Name": "banana-command",
									"HelpText": "banana-command"
								}
							]
						},
						"I-should-be-sorted-first": {
							"Location": "plugin-dir/plugin-1",
							"Version": {
								"Major": 1,
								"Minor": 2,
								"Build": 0
							},
							"Commands": [
								{
									"Name": "command-1",
									"HelpText": "some-command-1"
								},
								{
									"Name": "Better-command",
									"HelpText": "some-better-command"
								},
								{
									"Name": "command-2",
									"HelpText": "some-command-2"
								}
							]
						},
						"i-should-be-sorted-second": {
							"Location": "plugin-dir/plugin-2",
							"Version": {
								"Major": 1,
								"Minor": 0,
								"Build": 0
							},
							"Commands": [
								{
									"Name": "some-command",
									"HelpText": "some-command"
								},
								{
									"Name": "Some-other-command",
									"HelpText": "some-other-command"
								}
							]
						}
					}
				}`
			})

			It("displays the installed plugins in alphabetical order", func() {
				session := helpers.CF("plugins")
				Eventually(session).Should(Say("plugin name\\s+version\\s+command name\\s+command help"))
				Eventually(session).Should(Say("I-should-be-sorted-first\\s+1\\.2\\.0\\s+Better-command\\s+some-better-command"))
				Eventually(session).Should(Say("I-should-be-sorted-first\\s+1\\.2\\.0\\s+command-1\\s+some-command-1"))
				Eventually(session).Should(Say("I-should-be-sorted-first\\s+1\\.2\\.0\\s+command-2\\s+some-command-2"))
				Eventually(session).Should(Say("i-should-be-sorted-second\\s+1\\.0\\.0\\s+Some-other-command\\s+some-other-command"))
				Eventually(session).Should(Say("i-should-be-sorted-second\\s+1\\.0\\.0\\s+some-command\\s+some-command"))
				Eventually(session).Should(Say("sorted-third\\s+2\\.0\\.1\\s+banana-command\\s+banana-command"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when plugin version information is 0.0.0", func() {
			BeforeEach(func() {
				pluginConfig = `{
					"Plugins": {
						"some-plugin": {
							"Location": "plugin-dir/some-plugin",
							"Version": {
								"Major": 0,
								"Minor": 0,
								"Build": 0
							},
							"Commands": [
								{
									"Name": "banana-command",
									"HelpText": "banana-command"
								}
							]
						}
					}
				}`
			})

			It("displays N/A for the plugin's version", func() {
				session := helpers.CF("plugins")
				Eventually(session).Should(Say("some-plugin\\s+N/A"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when a plugin command has an alias", func() {
			BeforeEach(func() {
				pluginConfig = `{
					"Plugins": {
						"some-plugin": {
							"Location": "plugin-dir/some-plugin",
							"Version": {
								"Major": 1,
								"Minor": 0,
								"Build": 0
							},
							"Commands": [
								{
									"Name": "banana-command",
									"Alias": "bc",
									"HelpText": "banana-command"
								}
							]
						}
					}
				}`
			})

			It("displays the command name and it's alias", func() {
				session := helpers.CF("plugins")
				Eventually(session).Should(Say("some-plugin\\s+1\\.0\\.0\\s+banana-command, bc"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the --checksum flag is provided", func() {
			BeforeEach(func() {
				pluginConfig = fmt.Sprintf(`{
					"Plugins": {
						"some-plugin": {
							"Location": "%s/some-plugin",
							"Version": {
								"Major": 1,
								"Minor": 0,
								"Build": 0
							},
							"Commands": [
								{
									"Name": "banana-command",
									"HelpText": "banana-command"
								}
							]
						}
					}
				}`, pathToPlugins)
			})

			It("displays the sha1 value for each installed plugin", func() {
				err := ioutil.WriteFile(fmt.Sprintf("%s/some-plugin", pathToPlugins), []byte("some-text-to-sha"), 0600)
				Expect(err).NotTo(HaveOccurred())

				session := helpers.CF("plugins", "--checksum")
				Eventually(session).Should(Say("plugin name\\s+sha1"))
				Eventually(session).Should(Say("some-plugin\\s+88c2539b1dea4debfb127510c15c2aaebc3297a6"))
				Eventually(session).Should(Exit(0))
			})

			Context("when an error is encountered calculating the sha1 value", func() {
				It("displays N/A for the plugin's sha1", func() {
					err := ioutil.WriteFile(fmt.Sprintf("%s/some-plugin", pathToPlugins), []byte("some-text-to-sha"), 0000)
					Expect(err).NotTo(HaveOccurred())

					session := helpers.CF("plugins", "--checksum")
					Eventually(session).Should(Say("plugin name\\s+sha1"))
					Eventually(session).Should(Say("some-plugin\\s+N/A"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
