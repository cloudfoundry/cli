package common_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/commands/commandsfakes"
	. "code.cloudfoundry.org/cli/commands/v2/common"
	"code.cloudfoundry.org/cli/utils/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("New Clients", func() {
	var (
		binaryName string
		fakeConfig *commandsfakes.FakeConfig
		fakeUI     *ui.UI
	)

	BeforeEach(func() {
		binaryName = "faceman"
		fakeConfig = new(commandsfakes.FakeConfig)
		fakeUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())

		fakeConfig.BinaryNameReturns(binaryName)
	})

	Context("when the api endpoint is not set", func() {
		It("returns an error", func() {
			_, _, err := NewClients(fakeConfig, fakeUI)
			Expect(err).To(MatchError(NoAPISetError{
				BinaryName: binaryName,
			}))
		})
	})

	Context("when the DialTimeout is set", func() {
		BeforeEach(func() {
			if runtime.GOOS == "windows" {
				Skip("due to timing issues on windows")
			}
			fakeConfig.TargetReturns("https://potato.bananapants11122.co.uk")
			fakeConfig.DialTimeoutReturns(time.Nanosecond)
		})

		It("passes the value to the target", func() {
			_, _, err := NewClients(fakeConfig, fakeUI)
			Expect(err).To(MatchError("Get https://potato.bananapants11122.co.uk/v2/info: dial tcp: i/o timeout"))
		})
	})

	Context("when the targeting a CF fails", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("https://potato.bananapants11122.co.uk")
		})

		It("returns an error", func() {
			_, _, err := NewClients(fakeConfig, fakeUI)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when cf trace is true", func() {
		var server *Server
		BeforeEach(func() {
			server = NewTLSServer()
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/info"),
					RespondWith(http.StatusOK, "{}"),
				),
			)
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/apps"),
					RespondWith(http.StatusOK, "{}"),
				),
			)

			// Suppresses ginkgo server logs
			server.HTTPTestServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
			fakeConfig.TargetReturns(server.URL())
			fakeConfig.SkipSSLValidationReturns(true)
		})

		AfterEach(func() {
			server.Close()
		})

		Context("when outputting to terminal", func() {
			BeforeEach(func() {
				fakeConfig.VerboseReturns(true, "")
			})

			It("wraps the connection is a RequestLogger with TerminalDisplay", func() {
				client, _, err := NewClients(fakeConfig, fakeUI)
				Expect(err).ToNot(HaveOccurred())

				_, _, err = client.GetApplications(nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeUI.Out).To(Say("REQUEST"))
				Expect(fakeUI.Out).To(Say("RESPONSE"))
			})
		})

		Context("when outputting to a file", func() {
			var (
				tmpdir string
				tmpfn  string
			)

			BeforeEach(func() {
				var err error
				tmpdir, err = ioutil.TempDir("", "request_logger")
				tmpfn = filepath.Join(tmpdir, "log")
				Expect(err).ToNot(HaveOccurred())
				fakeConfig.VerboseReturns(true, tmpfn)
			})

			AfterEach(func() {
				os.RemoveAll(tmpdir)
			})

			It("wraps the connection is a RequestLogger with FileWriter", func() {
				client, _, err := NewClients(fakeConfig, fakeUI)
				Expect(err).ToNot(HaveOccurred())

				_, _, err = client.GetApplications(nil)
				Expect(err).ToNot(HaveOccurred())

				contents, err := ioutil.ReadFile(tmpfn)
				Expect(err).ToNot(HaveOccurred())

				output := string(contents)
				Expect(output).To(MatchRegexp("REQUEST"))
				Expect(output).To(MatchRegexp("RESPONSE"))
			})
		})
	})
})
