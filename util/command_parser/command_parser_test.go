package command_parser_test

import (
	"bytes"
	"io"

	"code.cloudfoundry.org/cli/v8/command/common"
	"code.cloudfoundry.org/cli/v8/util/command_parser"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
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
		pluginUI, err = ui.NewPluginUI(v3Config, io.Discard, io.Discard)
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

	Describe("repeated command parsing", func() {
		var (
			parser command_parser.CommandParser
			output bytes.Buffer
		)

		BeforeEach(func() {
			var err error
			parser, err = command_parser.NewCommandParser(v3Config)
			Expect(err).ToNot(HaveOccurred())
			pluginUI, err = ui.NewPluginUI(v3Config, &output, io.Discard)
			Expect(err).ToNot(HaveOccurred())
		})

		DescribeTable("does not carry checksum flags into the next command",
			func(firstArgs []string, expectedExitCode int) {
				exitCode, _ := parser.ParseCommandFromArgs(pluginUI, firstArgs)
				Expect(exitCode).To(Equal(expectedExitCode))
				output.Reset()

				exitCode, err := parser.ParseCommandFromArgs(pluginUI, []string{"plugins"})
				Expect(err).ToNot(HaveOccurred())
				Expect(exitCode).To(Equal(0))
				Expect(output.String()).To(ContainSubstring("command name"))
				Expect(output.String()).ToNot(ContainSubstring("sha1"))
			},
			Entry("after success", []string{"plugins", "--checksum"}, 0),
			Entry("after help", []string{"plugins", "--checksum", "--help"}, 0),
			Entry("after a flag error", []string{"plugins", "--checksum", "--not-a-real-flag"}, 1),
		)

		It("does not carry a repository string flag into the next command", func() {
			exitCode, _ := parser.ParseCommandFromArgs(pluginUI, []string{"repo-plugins", "-r", "missing-repository"})
			Expect(exitCode).To(Equal(1))
			output.Reset()

			exitCode, err := parser.ParseCommandFromArgs(pluginUI, []string{"repo-plugins"})
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
			Expect(output.String()).To(ContainSubstring("Getting plugins from all repositories"))
		})

		It("does not carry global verbosity into the next command", func() {
			exitCode, err := parser.ParseCommandFromArgs(pluginUI, []string{"-v", "plugins"})
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
			Expect(v3Config.Flags.Verbose).To(BeTrue())

			exitCode, err = parser.ParseCommandFromArgs(pluginUI, []string{"plugins"})
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
			Expect(v3Config.Flags.Verbose).To(BeFalse())
		})

		It("shows general help for empty arguments after a verbose command", func() {
			exitCode, err := parser.ParseCommandFromArgs(pluginUI, []string{"-v", "plugins"})
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
			output.Reset()

			exitCode, err = parser.ParseCommandFromArgs(pluginUI, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
			Expect(output.String()).To(ContainSubstring("Before getting started:"))
			Expect(v3Config.Flags.Verbose).To(BeFalse())
		})

		It("does not carry a help command argument into general help", func() {
			exitCode, err := parser.ParseCommandFromArgs(pluginUI, []string{"help", "plugins"})
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
			output.Reset()

			exitCode, err = parser.ParseCommandFromArgs(pluginUI, []string{"help"})
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
			Expect(output.String()).To(ContainSubstring("Before getting started:"))
		})

		It("does not carry flags between separate parser instances", func() {
			exitCode, err := parser.ParseCommandFromArgs(pluginUI, []string{"plugins", "--checksum"})
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
			output.Reset()

			secondParser, err := command_parser.NewCommandParser(v3Config)
			Expect(err).ToNot(HaveOccurred())
			exitCode, err = secondParser.ParseCommandFromArgs(pluginUI, []string{"plugins"})
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
			Expect(output.String()).ToNot(ContainSubstring("sha1"))
		})

		It("does not store destructive flags in shared command metadata", func() {
			exitCode, _ := parser.ParseCommandFromArgs(pluginUI, []string{"delete", "app1", "-f", "-r"})
			// No API is targeted, so neither invocation can delete anything.
			Expect(exitCode).To(Equal(1))

			exitCode, _ = parser.ParseCommandFromArgs(pluginUI, []string{"delete", "app2"})
			Expect(exitCode).To(Equal(1))
			Expect(common.Commands.Delete.Force).To(BeFalse())
			Expect(common.Commands.Delete.DeleteMappedRoutes).To(BeFalse())
		})
	})
})
