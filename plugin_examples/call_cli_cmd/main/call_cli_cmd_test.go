package main_test

import (
	"github.com/cloudfoundry/cli/plugin/fakes"
	. "github.com/cloudfoundry/cli/plugin_examples/call_cli_cmd/main"
	io_helpers "github.com/cloudfoundry/cli/testhelpers/io"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CallCliCmd", func() {
	Describe(".Run", func() {
		var fakeCliConnection *fakes.FakeCliConnection
		var callCliCommandPlugin *CliCmd

		BeforeEach(func() {
			fakeCliConnection = &fakes.FakeCliConnection{}
			callCliCommandPlugin = &CliCmd{}
		})

		It("calls the cli command that is passed as an argument", func() {
			io_helpers.CaptureOutput(func() {
				callCliCommandPlugin.Run(fakeCliConnection, []string{"cli-command", "plugins", "arg1"})
			})

			Expect(fakeCliConnection.CliCommandArgsForCall(0)[0]).To(Equal("plugins"))
			Expect(fakeCliConnection.CliCommandArgsForCall(0)[1]).To(Equal("arg1"))
		})

		It("ouputs the text returned by the cli command", func() {
			fakeCliConnection.CliCommandReturns([]string{"Hi", "Mom"}, nil)
			output := io_helpers.CaptureOutput(func() {
				callCliCommandPlugin.Run(fakeCliConnection, []string{"cli-command", "plugins", "arg1"})
			})

			Expect(output[1]).To(Equal("---------- Command output from the plugin ----------"))
			Expect(output[2]).To(Equal("# 0  value:  Hi"))
			Expect(output[3]).To(Equal("# 1  value:  Mom"))
			Expect(output[4]).To(Equal("----------              FIN               -----------"))
		})
	})
})
