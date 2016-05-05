package commands_test

import (
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/configuration/pluginconfig"
	"github.com/cloudfoundry/cli/cf/configuration/pluginconfig/pluginconfigfakes"
	"github.com/cloudfoundry/cli/commandsloader"
	"github.com/cloudfoundry/cli/plugin"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"

	"github.com/cloudfoundry/cli/cf/terminal/terminalfakes"
	"github.com/cloudfoundry/cli/flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Help", func() {

	commandsloader.Load()

	var (
		fakeFactory *testreq.FakeReqFactory
		fakeUI      *terminalfakes.FakeUI
		fakeConfig  *pluginconfigfakes.FakePluginConfiguration
		deps        commandregistry.Dependency

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
		fakeFactory = &testreq.FakeReqFactory{}
	})

	AfterEach(func() {
		buffer.Close()
	})

	Context("when no argument is provided", func() {
		It("prints the main help menu of the 'cf' app", func() {
			flagContext.Parse()
			cmd.Execute(flagContext)

			Eventually(buffer.Contents()).Should(ContainSubstring("A command line tool to interact with Cloud Foundry"))
			Eventually(buffer).Should(gbytes.Say("CF_TRACE=true"))
		})
	})

	Context("when a command name is provided as an argument", func() {
		Context("When the command exists", func() {
			It("prints the usage help for the command", func() {
				flagContext.Parse("target")
				cmd.Execute(flagContext)

				Expect(fakeUI.SayCallCount()).To(Equal(1))
				output, _ := fakeUI.SayArgsForCall(0)
				Expect(output).To(ContainSubstring("target - Set or view the targeted org or space"))
			})
		})

		Context("When the command does not exists", func() {
			It("prints the usage help for the command", func() {
				flagContext.Parse("bad-command")
				cmd.Execute(flagContext)

				Expect(fakeUI.FailedCallCount()).To(Equal(1))
				output, _ := fakeUI.FailedArgsForCall(0)
				Expect(output).To(ContainSubstring("'bad-command' is not a registered command. See 'cf help'"))
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
				cmd.Execute(flagContext)

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
				cmd.Execute(flagContext)

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
