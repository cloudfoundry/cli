package commands_test

import (
	"strings"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands"
	"code.cloudfoundry.org/cli/cf/commandsloader"
	"code.cloudfoundry.org/cli/cf/configuration/pluginconfig"
	"code.cloudfoundry.org/cli/cf/configuration/pluginconfig/pluginconfigfakes"
	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/plugin"

	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Help", func() {

	commandsloader.Load()

	var (
		fakeUI     *terminalfakes.FakeUI
		fakeConfig *pluginconfigfakes.FakePluginConfiguration
		deps       commandregistry.Dependency

		cmd         *commands.Help
		flagContext flags.FlagContext
		buffer      *gbytes.Buffer
	)

	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
		fakeUI = new(terminalfakes.FakeUI)
		fakeUI.WriterReturns(buffer)

		fakeConfig = new(pluginconfigfakes.FakePluginConfiguration)

		deps = commandregistry.Dependency{
			UI:           fakeUI,
			PluginConfig: fakeConfig,
		}

		cmd = &commands.Help{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
	})

	AfterEach(func() {
		buffer.Close()
	})

	Context("when no argument is provided", func() {
		It("prints the main help menu of the 'cf' app", func() {
			flagContext.Parse()
			err := cmd.Execute(flagContext)
			Expect(err).NotTo(HaveOccurred())

			Eventually(buffer.Contents).Should(ContainSubstring("A command line tool to interact with Cloud Foundry"))
			Eventually(buffer).Should(gbytes.Say("CF_TRACE=true"))
		})
	})

	Context("when a command name is provided as an argument", func() {
		Context("When the command exists", func() {
			It("prints the usage help for the command", func() {
				flagContext.Parse("target")
				err := cmd.Execute(flagContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeUI.SayCallCount()).To(Equal(1))
				output, _ := fakeUI.SayArgsForCall(0)
				Expect(output).To(ContainSubstring("target - Set or view the targeted org or space"))
			})

			Context("i18n translations", func() {
				var originalT func(string, ...interface{}) string

				BeforeEach(func() {
					originalT = i18n.T
				})

				AfterEach(func() {
					i18n.T = originalT
				})

				It("includes ':' in caption translation strings for language like French to be translated correctly", func() {
					nameCaption := "NAME:"
					aliasCaption := "ALIAS:"
					usageCaption := "USAGE:"
					optionsCaption := "OPTIONS:"
					captionCheckCount := 0

					i18n.T = func(translationID string, args ...interface{}) string {
						if strings.HasPrefix(translationID, "NAME") {
							Expect(translationID).To(Equal(nameCaption))
							captionCheckCount += 1
						} else if strings.HasPrefix(translationID, "ALIAS") {
							Expect(translationID).To(Equal(aliasCaption))
							captionCheckCount += 1
						} else if strings.HasPrefix(translationID, "USAGE") {
							Expect(translationID).To(Equal(usageCaption))
							captionCheckCount += 1
						} else if strings.HasPrefix(translationID, "OPTIONS") {
							Expect(translationID).To(Equal(optionsCaption))
							captionCheckCount += 1
						}

						return translationID
					}

					flagContext.Parse("target")
					err := cmd.Execute(flagContext)
					Expect(err).NotTo(HaveOccurred())

					Expect(captionCheckCount).To(Equal(4))
				})
			})
		})

		Context("When the command does not exists", func() {
			It("prints the usage help for the command", func() {
				flagContext.Parse("bad-command")
				err := cmd.Execute(flagContext)
				Expect(err.Error()).To(Equal("'bad-command' is not a registered command. See 'cf help -a'"))
			})
		})
	})

	Context("when a command provided is a plugin command", func() {
		BeforeEach(func() {
			m := make(map[string]pluginconfig.PluginMetadata)
			m["fakePlugin"] = pluginconfig.PluginMetadata{
				Commands: []plugin.Command{
					{
						Name:     "fakePluginCmd1",
						Alias:    "fpc1",
						HelpText: "help text here",
						UsageDetails: plugin.Usage{
							Usage: "Usage for fpc1",
							Options: map[string]string{
								"f": "test flag",
							},
						},
					},
				},
			}

			fakeConfig.PluginsReturns(m)
		})

		Context("command is a plugin command name", func() {
			It("prints the usage help for the command", func() {
				flagContext.Parse("fakePluginCmd1")
				err := cmd.Execute(flagContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeUI.SayCallCount()).To(Equal(1))
				output, _ := fakeUI.SayArgsForCall(0)
				Expect(output).To(ContainSubstring("fakePluginCmd1"))
				Expect(output).To(ContainSubstring("help text here"))
				Expect(output).To(ContainSubstring("ALIAS"))
				Expect(output).To(ContainSubstring("fpc1"))
				Expect(output).To(ContainSubstring("USAGE"))
				Expect(output).To(ContainSubstring("Usage for fpc1"))
				Expect(output).To(ContainSubstring("OPTIONS"))
				Expect(output).To(ContainSubstring("-f"))
				Expect(output).To(ContainSubstring("test flag"))
			})
		})

		Context("command is a plugin command alias", func() {
			It("prints the usage help for the command alias", func() {
				flagContext.Parse("fpc1")
				err := cmd.Execute(flagContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeUI.SayCallCount()).To(Equal(1))
				output, _ := fakeUI.SayArgsForCall(0)
				Expect(output).To(ContainSubstring("fakePluginCmd1"))
				Expect(output).To(ContainSubstring("help text here"))
				Expect(output).To(ContainSubstring("ALIAS"))
				Expect(output).To(ContainSubstring("fpc1"))
				Expect(output).To(ContainSubstring("USAGE"))
				Expect(output).To(ContainSubstring("Usage for fpc1"))
				Expect(output).To(ContainSubstring("OPTIONS"))
				Expect(output).To(ContainSubstring("-f"))
				Expect(output).To(ContainSubstring("test flag"))
			})
		})

	})
})
