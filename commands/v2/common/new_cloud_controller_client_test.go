package common_test

import (
	"bytes"
	"log"
	"net/http"

	"code.cloudfoundry.org/cli/commands/commandsfakes"
	. "code.cloudfoundry.org/cli/commands/v2/common"
	"code.cloudfoundry.org/cli/utils/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("New Cloud Controller Client", func() {
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
			_, err := NewCloudControllerClient(fakeConfig, fakeUI)
			Expect(err).To(MatchError(NoAPISetError{
				BinaryName: binaryName,
			}))
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
			fakeConfig.VerboseReturns(true, "")
		})

		AfterEach(func() {
			server.Close()
		})

		It("wraps the connection is a RequestLogger", func() {
			client, err := NewCloudControllerClient(fakeConfig, fakeUI)
			Expect(err).ToNot(HaveOccurred())

			_, _, err = client.GetApplications(nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeUI.Out).To(Say("REQUEST"))
			Expect(fakeUI.Out).To(Say("RESPONSE"))
		})
	})
})
