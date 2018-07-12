package shared_test

import (
	"net/http"
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
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
			_, _, err := NewClients(fakeConfig, testUI, true)
			Expect(err).To(MatchError(translatableerror.NoAPISetError{
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

		Context("when the error is a cloud controller request error", func() {
			BeforeEach(func() {
				fakeConfig.TargetReturns("https://127.0.0.1:9876")
			})

			It("returns a command api request error", func() {
				_, _, err := NewClients(fakeConfig, testUI, true)
				Expect(err).To(MatchError(ContainSubstring("dial")))
			})
		})

		Context("when the error is a cloud controller api not found error", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/"),
						RespondWith(http.StatusNotFound, "something which is not json"),
					),
				)
			})

			It("returns a command api not found error", func() {
				_, _, err := NewClients(fakeConfig, testUI, true)
				Expect(err).To(MatchError(ccerror.APINotFoundError{URL: server.URL()}))
			})
		})

		Context("when the error is a V3UnexpectedResponseError and the status code is 404", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/"),
						RespondWith(http.StatusNotFound, "{}"),
					),
				)
			})

			It("returns a V3APIDoesNotExistError", func() {
				_, _, err := NewClients(fakeConfig, testUI, true)
				expectedErr := ccerror.V3UnexpectedResponseError{ResponseCode: http.StatusNotFound}
				Expect(err).To(MatchError(expectedErr))
			})
		})

		Context("when the error is generic and the body is valid json", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/"),
						RespondWith(http.StatusTeapot, `{ "some-error": "invalid" }`),
					),
				)
			})

			It("returns a V3UnexpectedResponseError", func() {
				_, _, err := NewClients(fakeConfig, testUI, true)
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{ResponseCode: http.StatusTeapot}))
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
			_, _, err := NewClients(fakeConfig, testUI, true)
			Expect(err.Error()).To(MatchRegexp("timeout"))
		})
	})

	Context("when not targetting", func() {
		It("does not target and returns no UAA client", func() {
			ccClient, uaaClient, err := NewClients(fakeConfig, testUI, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ccClient).ToNot(BeNil())
			Expect(uaaClient).To(BeNil())
			Expect(fakeConfig.SkipSSLValidationCallCount()).To(Equal(0))
		})
	})
})
