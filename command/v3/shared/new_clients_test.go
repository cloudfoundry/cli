package shared_test

import (
	"net/http"
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("New Clients", func() {
	var (
		binaryName string
		fakeConfig *commandfakes.FakeConfig
		testUI     *ui.UI
	)

	BeforeEach(func() {
		binaryName = "faceman"
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns(binaryName)

		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
	})

	Context("when the api endpoint is not set", func() {
		It("returns the NoAPISetError", func() {
			_, err := NewClients(fakeConfig, testUI, true)
			Expect(err).To(MatchError(command.NoAPISetError{
				BinaryName: binaryName,
			}))
		})
	})

	Context("when hitting the target returns an error", func() {
		var server *Server
		BeforeEach(func() {
			server = NewTLSServer()

			fakeConfig.TargetReturns(server.URL())
			fakeConfig.SkipSSLValidationReturns(true)
		})

		AfterEach(func() {
			server.Close()
		})

		Context("that is a cloud controller request error", func() {
			BeforeEach(func() {
				fakeConfig.TargetReturns("https://127.0.0.1:9876")
			})

			It("returns a command api request error", func() {
				_, err := NewClients(fakeConfig, testUI, true)
				Expect(err).To(MatchError(ContainSubstring("Request error:")))
			})

		})

		Context("that is a cloud controller api not found error", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/"),
						RespondWith(http.StatusNotFound, "{}"),
					),
				)
			})

			It("returns a command api not found error", func() {
				_, err := NewClients(fakeConfig, testUI, true)
				Expect(err).To(MatchError(command.APINotFoundError{URL: server.URL()}))
			})
		})

		Context("that is another error", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/"),
						RespondWith(http.StatusOK, "invalid json"),
					),
				)
			})

			{
				_, err := NewClients(fakeConfig, testUI, true)
				Expect(err.Error()).To(MatchRegexp("Note that this command requires CF API version 3.0.0+."))
			})
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
			_, err := NewClients(fakeConfig, testUI, true)
			Expect(err.Error()).To(MatchRegexp("TIP: If you are behind a firewall"))
		})
	})

	Context("when not targetting", func() {
		It("does not target and returns no UAA client", func() {
			ccClient, err := NewClients(fakeConfig, testUI, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ccClient).ToNot(BeNil())
			Expect(fakeConfig.SkipSSLValidationCallCount()).To(Equal(0))
		})
	})
})
