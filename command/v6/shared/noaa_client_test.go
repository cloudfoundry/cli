package shared_test

import (
	"reflect"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command/commandfakes"

	"code.cloudfoundry.org/cli/command/v6/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewNOAAClient", func() {
	When("a new NOAA client is created", func() {
		var (
			fakeConfig *commandfakes.FakeConfig
			testUI     *commandfakes.FakeUI
			uaaClient  *uaa.Client
		)

		BeforeEach(func() {
			fakeConfig = new(commandfakes.FakeConfig)
			testUI = new(commandfakes.FakeUI)
			uaaClient = new(uaa.Client)
		})

		It("always sets a token refresher", func() {
			// We need to be sure that the token refresher is set on the NOAA client.
			// However, this struct is in a library we don't control and the field iss
			// private.  Testing this functionality via an integration test would
			// require the test to both wait for an OAuth token to expire and kill a
			// TCP connection.  Making a reflection-free unit test would require
			// changing everywhere we pass around a consumer.Consumer to instead use
			// an interface.
			client := shared.NewNOAAClient("some-url", fakeConfig, uaaClient, testUI)
			// This warning is expected. This access to ValueOf(*client) is not
			// threadsafe, but we only have this local variable.
			v := reflect.ValueOf(*client) // nolint: vet
			refresher := v.FieldByName("tokenRefresher")
			Expect(refresher.IsNil()).To(BeFalse(), "NewNOAAClient failed to call RefreshTokenFrom on NOAA client")
		})
	})
})
