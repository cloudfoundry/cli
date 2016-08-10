package v2_test

import (
	"code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/commands/flags"
	. "code.cloudfoundry.org/cli/commands/v2"
	"code.cloudfoundry.org/cli/commands/v2/customv2fakes"
	"code.cloudfoundry.org/cli/commands/v2/v2fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Help Command", func() {
	var (
		fakeUI    *customv2fakes.FakeUI
		fakeActor *v2fakes.FakeHelpActor
		cmd       HelpCommand
	)

	BeforeEach(func() {
		fakeUI = customv2fakes.NewFakeUI(true)
		fakeActor = new(v2fakes.FakeHelpActor)

		cmd = HelpCommand{
			UI:    fakeUI,
			Actor: fakeActor,
		}
	})

	Context("providing help for a specific command", func() {
		BeforeEach(func() {
			cmd.OptionalArgs = flags.CommandName{
				CommandName: "help",
			}

			commandInfo := v2actions.CommandInfo{
				Name:        "help",
				Description: "Show help",
				Usage:       "cf help [COMMAND]",
				Alias:       "h",
			}
			fakeActor.GetCommandInfoReturns(commandInfo, nil)
		})

		It("displays the name for help", func() {
			err := cmd.Execute(nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeUI.Out).To(Say("NAME:"))
			Expect(fakeUI.Out).To(Say("    help - Show help"))

			Expect(fakeActor.GetCommandInfoCallCount()).To(Equal(1))
			_, commandName := fakeActor.GetCommandInfoArgsForCall(0)
			Expect(commandName).To(Equal("help"))
		})

		It("displays the usage for help", func() {
			err := cmd.Execute(nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeUI.Out).To(Say("NAME:"))
			Expect(fakeUI.Out).To(Say("USAGE:"))
			Expect(fakeUI.Out).To(Say("    cf help \\[COMMAND\\]"))
		})

		Describe("aliases", func() {
			Context("when the command has an alias", func() {
				It("displays the alias for help", func() {
					err := cmd.Execute(nil)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeUI.Out).To(Say("USAGE:"))
					Expect(fakeUI.Out).To(Say("ALIAS:"))
					Expect(fakeUI.Out).To(Say("    h"))
				})
			})

			Context("when the command does not have an alias", func() {
				BeforeEach(func() {
					cmd.OptionalArgs = flags.CommandName{
						CommandName: "app",
					}

					commandInfo := v2actions.CommandInfo{
						Name: "app",
					}
					fakeActor.GetCommandInfoReturns(commandInfo, nil)
				})

				It("no alias is displayed", func() {
					err := cmd.Execute(nil)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeUI.Out).ToNot(Say("ALIAS:"))
				})
			})
		})

		Describe("options", func() {
			Context("when the command has options", func() {
				BeforeEach(func() {
					cmd.OptionalArgs = flags.CommandName{
						CommandName: "push",
					}
					commandInfo := v2actions.CommandInfo{
						Name: "push",
						Flags: []v2actions.CommandFlag{
							{
								Long:        "no-hostname",
								Description: "Map the root domain to this app",
							},
							{
								Short:       "b",
								Description: "Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'",
							},
							{
								Long:        "hostname",
								Short:       "n",
								Description: "Hostname (e.g. my-subdomain)",
							},
						},
					}
					fakeActor.GetCommandInfoReturns(commandInfo, nil)
				})

				Context("only has a long option", func() {
					It("displays the options for app", func() {
						err := cmd.Execute(nil)
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeUI.Out).To(Say("USAGE:"))
						Expect(fakeUI.Out).To(Say("OPTIONS:"))
						Expect(fakeUI.Out).To(Say("--no-hostname\\s+Map the root domain to this app"))
					})
				})

				Context("only has a short option", func() {
					It("displays the options for app", func() {
						err := cmd.Execute(nil)
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeUI.Out).To(Say("USAGE:"))
						Expect(fakeUI.Out).To(Say("OPTIONS:"))
						Expect(fakeUI.Out).To(Say("-b\\s+Custom buildpack by name \\(e.g. my-buildpack\\) or Git URL \\(e.g. 'https://github.com/cloudfoundry/java-buildpack.git'\\) or Git URL with a branch or tag \\(e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag\\). To use built-in buildpacks only, specify 'default' or 'null'"))
					})
				})

				Context("has long and short options", func() {
					It("displays the options for app", func() {
						err := cmd.Execute(nil)
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeUI.Out).To(Say("USAGE:"))
						Expect(fakeUI.Out).To(Say("OPTIONS:"))
						Expect(fakeUI.Out).To(Say("--hostname, -n\\s+Hostname \\(e.g. my-subdomain\\)"))
					})
				})

				Context("has hidden options", func() {
					It("does not display the hidden option", func() {
						err := cmd.Execute(nil)
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeUI.Out).ToNot(Say("--app-ports"))
					})
				})
			})
		})
	})

	PContext("providing help for all commands", func() {

	})
})
