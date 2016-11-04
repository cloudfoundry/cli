package ccv2_test

import (
	"fmt"
	"net/http"
	"runtime"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/ccv2fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Cloud Controller Client", func() {
	var (
		client *Client
	)

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("WrapConnection", func() {
		var fakeConnectionWrapper *ccv2fakes.FakeConnectionWrapper

		BeforeEach(func() {
			fakeConnectionWrapper = new(ccv2fakes.FakeConnectionWrapper)
			fakeConnectionWrapper.WrapReturns(fakeConnectionWrapper)
		})

		It("wraps the existing connection in the provided wrapper", func() {
			client.WrapConnection(fakeConnectionWrapper)
			Expect(fakeConnectionWrapper.WrapCallCount()).To(Equal(1))

			client.DeleteServiceBinding("does-not-matter")
			Expect(fakeConnectionWrapper.MakeCallCount()).To(Equal(1))
		})
	})

	Describe("User Agent", func() {
		BeforeEach(func() {
			expectedUserAgent := fmt.Sprintf("CF CLI API V2 Test/Unknown (%s; %s %s)", runtime.Version(), runtime.GOARCH, runtime.GOOS)
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/apps"),
					VerifyHeaderKV("User-Agent", expectedUserAgent),
					RespondWith(http.StatusOK, "{}"),
				),
			)
		})

		It("adds a user agent header", func() {
			client.GetApplications(nil)
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})
	})
})
