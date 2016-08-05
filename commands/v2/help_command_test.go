package v2_test

import (
	"code.cloudfoundry.org/cli/commands/flags"
	. "code.cloudfoundry.org/cli/commands/v2"
	"code.cloudfoundry.org/cli/commands/v2/customv2fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Help Command", func() {
	var (
		fakeUI *customv2fakes.FakeUI

		cmd HelpCommand
	)

	BeforeEach(func() {
		fakeUI = customv2fakes.NewFakeUI(true)

		cmd = HelpCommand{
			UI: fakeUI,
		}
	})

	Context("providing help for a command", func() {
		BeforeEach(func() {
			cmd.OptionalArgs = flags.CommandName{
				CommandName: "heLp", //Help cased incorrectly on purpose
			}
		})

		It("displays the name for help", func() {
			err := cmd.Execute(nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeUI.Out).To(Say("NAME:"))
			Expect(fakeUI.Out).To(Say("\thelp - Show help"))
		})

		It("displays the usage for help", func() {
			err := cmd.Execute(nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeUI.Out).To(Say("NAME:"))
			Expect(fakeUI.Out).To(Say("USAGE:"))
			Expect(fakeUI.Out).To(Say("\tcf help \\[COMMAND\\]"))
		})

		Describe("aliases", func() {
			Context("when the command has an alias", func() {
				It("displays the alias for help", func() {
					err := cmd.Execute(nil)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeUI.Out).To(Say("USAGE:"))
					Expect(fakeUI.Out).To(Say("ALIAS:"))
					Expect(fakeUI.Out).To(Say("\th"))
				})
			})

			Context("when the command does not have an alias", func() {
				BeforeEach(func() {
					cmd.OptionalArgs = flags.CommandName{
						CommandName: "app",
					}
				})

				It("no alias is displayed", func() {
					err := cmd.Execute(nil)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeUI.Out).ToNot(Say("ALIAS:"))
				})
			})
		})

		FDescribe("options", func() {
			Context("when the command has options", func() {
				BeforeEach(func() {
					cmd.OptionalArgs = flags.CommandName{
						CommandName: "push",
					}
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
