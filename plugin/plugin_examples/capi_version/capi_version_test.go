package main_test

import (
	"errors"
	"io"

	"code.cloudfoundry.org/cli/plugin"
	main "code.cloudfoundry.org/cli/plugin/plugin_examples/capi_version"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("CAPI version plugin", func() {

	var (
		versionPlugin plugin.Plugin
		conn          *pluginfakes.FakeCliConnection
		outWriter     io.WriteCloser
		errWriter     io.WriteCloser
		stdout        *gbytes.Buffer
		stderr        *gbytes.Buffer
	)

	BeforeEach(func() {
		var (
			outReader io.Reader
			errReader io.Reader
		)
		outReader, outWriter = io.Pipe()
		stdout = gbytes.BufferReader(outReader)
		errReader, errWriter = io.Pipe()
		stderr = gbytes.BufferReader(errReader)

		versionPlugin = main.NewCapiVersionPlugin(outWriter, errWriter)
		conn = new(pluginfakes.FakeCliConnection)

		conn.CliCommandWithoutTerminalOutputReturns([]string{
			`{`,
			`   "name": "Some Cloud Foundry distribution",`,
			`   "build": "1.2.3-build.234",`,
			` 	"support": "https://example.org",`,
			`	  "description": "https://example.org/support/documentation",`,
			`	  "some_other_field": "some-other-value"`,
			`}`,
		}, nil)
	})

	JustBeforeEach(func() {
		versionPlugin.Run(conn, []string{"capi-version"})
	})

	AfterEach(func() {
		stdout.Close()
		stderr.Close()
	})

	It("retrieves version information from the /v2/info endpoint", func() {
		Expect(conn.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))
		Expect(conn.CliCommandWithoutTerminalOutputArgsForCall(0)).To(Equal([]string{"curl", "/v2/info"}))
	})

	It("prints the version details to stdout", func() {
		Eventually(stdout).Should(gbytes.Say("name: Some Cloud Foundry distribution"))
		Eventually(stdout).Should(gbytes.Say("build: 1.2.3-build.234"))
		Eventually(stdout).Should(gbytes.Say("support address: https://example.org"))
		Eventually(stdout).Should(gbytes.Say("description: https://example.org/support/documentation"))
	})

	When("there is an error calling the plugin API", func() {
		BeforeEach(func() {
			conn.CliCommandWithoutTerminalOutputReturns(nil, errors.New("some error"))
		})
		It("prints an error to stderr", func() {
			Eventually(stderr).Should(gbytes.Say("Problem retrieving version information"))
		})
	})

	When("the response from the server is malformed", func() {
		BeforeEach(func() {
			conn.CliCommandWithoutTerminalOutputReturns([]string{"{malformed"}, nil)
		})
		It("prints an error to stderr", func() {
			Eventually(stderr).Should(gbytes.Say("Unexpected response from server"))
		})
	})
})
