package shared_test

import (
	"net/http"
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("New Clients", func() {
	var (
		binaryName    string
		fakeConfig    *commandfakes.FakeConfig
		testUI        *ui.UI
		fakeUaaClient *uaa.Client
	)

	BeforeEach(func() {
		binaryName = "faceman"
		fakeConfig = new(commandfakes.FakeConfig)
		fakeUaaClient = &uaa.Client{}
		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())

		fakeConfig.BinaryNameReturns(binaryName)
	})

	Describe("NewClients", func() {
		When("the api endpoint is not set", func() {
			It("returns an error", func() {
				_, _, err := NewClients(fakeConfig, testUI, true)
				Expect(err).To(MatchError(translatableerror.NoAPISetError{
					BinaryName: binaryName,
				}))
			})
		})

		When("the DialTimeout is set", func() {
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

		When("the targeting a CF fails", func() {
			BeforeEach(func() {
				fakeConfig.TargetReturns("https://potato.bananapants11122.co.uk")
			})

			It("returns an error", func() {
				_, _, err := NewClients(fakeConfig, testUI, true)
				Expect(err).To(HaveOccurred())
			})
		})

		When("the targeted CF is older than the minimum supported version", func() {
			var server *Server

			BeforeEach(func() {
				server = NewTLSServer()

				fakeConfig.TargetReturns(server.URL())
				fakeConfig.SkipSSLValidationReturns(true)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/info"),
						RespondWith(http.StatusOK, `{ "api_version": "2.68.0" }`),
					),
				)
			})

			AfterEach(func() {
				server.Close()
			})

			It("outputs a warning", func() {
				NewClients(fakeConfig, testUI, true)

				Expect(testUI.Err).To(Say("Your API version is no longer supported. Upgrade to a newer version of the API"))
			})
		})

		When("not targetting", func() {
			It("does not target and returns no UAA client", func() {
				ccClient, uaaClient, err := NewClients(fakeConfig, testUI, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(ccClient).ToNot(BeNil())
				Expect(uaaClient).To(BeNil())
				Expect(fakeConfig.SkipSSLValidationCallCount()).To(Equal(0))
			})
		})
	})

	Describe("NewRouterClient", func() {
		It("should return a new router client", func() {
			routerClient, err := NewRouterClient(fakeConfig, testUI, fakeUaaClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(routerClient).ToNot(BeNil())
		})

		It("reads the app name and app version for its own config", func() {
			_, _ = NewRouterClient(fakeConfig, testUI, fakeUaaClient)
			Expect(fakeConfig.BinaryNameCallCount()).To(Equal(1))
			Expect(fakeConfig.BinaryVersionCallCount()).To(Equal(1))
		})
	})

})
