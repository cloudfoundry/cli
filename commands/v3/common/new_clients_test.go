package common_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/commands/commandsfakes"
	. "code.cloudfoundry.org/cli/commands/v3/common"
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
		fakeConfig.BinaryNameReturns(binaryName)

		fakeUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
	})

	Context("when the api endpoint is not set", func() {
		It("returns an error", func() {
			_, err := NewClients(fakeConfig, fakeUI)
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
			_, err := NewClients(fakeConfig, fakeUI)
			Expect(err).To(MatchError("Get https://potato.bananapants11122.co.uk: dial tcp: i/o timeout"))
		})
	})

	Context("when cf trace is true", func() {
		var server *Server
		BeforeEach(func() {
			server = NewTLSServer()
			serverURL := server.URL()
			rootResponse := fmt.Sprintf(`{
				"links": {
					"self": {
						"href": "%s"
					},
					"cloud_controller_v2": {
						"href": "%s/v2",
						"meta": {
							"version": "2.64.0"
						}
					},
					"cloud_controller_v3": {
						"href": "%s/v3",
						"meta": {
							"version": "3.0.0-alpha.5"
						}
					}
				}
			}`, serverURL, serverURL, serverURL)

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/"),
					RespondWith(
						http.StatusOK,
						rootResponse,
						http.Header{"X-Cf-Warnings": {"warning 1"}}),
				),
			)

			v3Response := fmt.Sprintf(`{
				"links": {
					"self": {
						"href": "%s/v3"
					},
					"tasks": {
						"href": "%s/v3/tasks"
					},
					"uaa": {
						"href": "https://uaa.bosh-lite.com"
					}
				}
			}`, serverURL, serverURL)

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v3"),
					RespondWith(
						http.StatusOK,
						v3Response,
						http.Header{"X-Cf-Warnings": {"warning 2"}}),
				),
			)

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v3/apps"),
					RespondWith(
						http.StatusOK,
						"{}",
						http.Header{"X-Cf-Warnings": {"warning 2"}}),
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
				client, err := NewClients(fakeConfig, fakeUI)
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
				client, err := NewClients(fakeConfig, fakeUI)
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
