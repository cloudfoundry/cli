package command_parser_test

import (
	"code.cloudfoundry.org/cli/util/command_parser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command 'Parser'", func() {

	It("returns an unknown command error", func() {
		status := command_parser.ParseCommandFromArgs([]string{"howdy"})
		Expect(status).To(Equal(-666))
	})
})
