package command_parser_test

import (
	"io/ioutil"

	"code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/util/command_parser"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command 'Parser'", func() {
	var (
		pluginUI *ui.UI
		v3Config *configv3.Config
	)
	BeforeEach(func() {
		var err error
		v3Config = new(configv3.Config)
		pluginUI, err = ui.NewPluginUI(v3Config, ioutil.Discard, ioutil.Discard)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("unknown command handling", func() {
		var (
			exitCode int
			err      error
		)

		BeforeEach(func() {
			parser, newErr := command_parser.NewCommandParser(v3Config)
			Expect(newErr).ToNot(HaveOccurred())
			exitCode, err = parser.ParseCommandFromArgs(pluginUI, []string{"howdy"})
		})

		It("returns an unknown command error with the command name", func() {
			unknownCommandErr := err.(command_parser.UnknownCommandError)
			Expect(unknownCommandErr.CommandName).To(Equal("howdy"))
			Expect(exitCode).To(Equal(0))
		})
	})

	Describe("the verbose flag", func() {
		var parser command_parser.CommandParser

		BeforeEach(func() {
			// Needed because the command-table is a singleton
			// and the absence of -v relies on the default value of
			// common.Commands.VerboseOrVersion to be false
			common.Commands.VerboseOrVersion = false
			var err error

			parser, err = command_parser.NewCommandParser(v3Config)
			Expect(err).ToNot(HaveOccurred())
		})

		It("sets the verbose/version flag", func() {
			exitCode, err := parser.ParseCommandFromArgs(pluginUI, []string{"-v", "help"})
			Expect(exitCode).To(Equal(0))
			Expect(err).ToNot(HaveOccurred())
			Expect(parser.Config.Flags).To(Equal(configv3.FlagOverride{Verbose: true}))
		})

		It("sets the verbose/version flag after the command-name", func() {
			exitCode, err := parser.ParseCommandFromArgs(pluginUI, []string{"help", "-v"})
			Expect(exitCode).To(Equal(0))
			Expect(err).ToNot(HaveOccurred())
			Expect(parser.Config.Flags).To(Equal(configv3.FlagOverride{Verbose: true}))
		})

		It("doesn't turn verbose on by default", func() {
			exitCode, err := parser.ParseCommandFromArgs(pluginUI, []string{"help"})
			Expect(exitCode).To(Equal(0))
			Expect(err).ToNot(HaveOccurred())
			Expect(parser.Config.Flags).To(Equal(configv3.FlagOverride{Verbose: false}))
		})

	})
})
