package command_parser_test

import (
	"io"
	"io/ioutil"

	"code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/util/command_parser"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command 'Parser'", func() {

	It("returns an unknown command error", func() {
		parser, err := command_parser.NewCommandParser()
		Expect(err).ToNot(HaveOccurred())
		status := parser.ParseCommandFromArgs([]string{"howdy"})
		Expect(status).To(Equal(-666))
	})

	Describe("the verbose flag", func() {
		var stdout, outBuffer, errBuffer io.Writer

		BeforeEach(func() {
			stdout = ioutil.Discard
			outBuffer = ioutil.Discard
			errBuffer = ioutil.Discard
			// Needed because the command-table is a singleton
			// and the absence of -v relies on the default value of
			// common.Commands.VerboseOrVersion to be false
			common.Commands.VerboseOrVersion = false
		})

		It("sets the verbose/version flag", func() {
			parser, err := command_parser.NewCommandParserForPlugins(stdout, outBuffer, errBuffer)
			Expect(err).ToNot(HaveOccurred())
			status := parser.ParseCommandFromArgs([]string{"-v", "help"})
			Expect(status).To(Equal(0))
			Expect(parser.Config.Flags).To(Equal(configv3.FlagOverride{Verbose: true}))
		})

		It("sets the verbose/version flag after the command-name", func() {
			parser, err := command_parser.NewCommandParserForPlugins(stdout, outBuffer, errBuffer)
			Expect(err).ToNot(HaveOccurred())
			status := parser.ParseCommandFromArgs([]string{"help", "-v"})
			Expect(status).To(Equal(0))
			Expect(parser.Config.Flags).To(Equal(configv3.FlagOverride{Verbose: true}))
		})

		It("doesn't turn verbose on by default", func() {
			parser, err := command_parser.NewCommandParserForPlugins(stdout, outBuffer, errBuffer)
			Expect(err).ToNot(HaveOccurred())
			status := parser.ParseCommandFromArgs([]string{"help"})
			Expect(status).To(Equal(0))
			Expect(parser.Config.Flags).To(Equal(configv3.FlagOverride{Verbose: false}))
		})

	})
})
