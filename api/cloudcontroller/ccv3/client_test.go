package ccv3_test

import (
	"fmt"
	"net/http"
	"runtime"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"

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
		var fakeConnectionWrapper *ccv3fakes.FakeConnectionWrapper

		BeforeEach(func() {
			fakeConnectionWrapper = new(ccv3fakes.FakeConnectionWrapper)
			fakeConnectionWrapper.WrapReturns(fakeConnectionWrapper)
		})

		It("wraps the existing connection in the provided wrapper", func() {
			client.WrapConnection(fakeConnectionWrapper)
			Expect(fakeConnectionWrapper.WrapCallCount()).To(Equal(1))

			client.Info()
			Expect(fakeConnectionWrapper.MakeCallCount()).To(Equal(2))
		})
	})

	Describe("User Agent", func() {
		BeforeEach(func() {
			expectedUserAgent := fmt.Sprintf("CF CLI API V3 Test/Unknown (%s; %s %s)", runtime.Version(), runtime.GOARCH, runtime.GOOS)
			rootResponse := fmt.Sprintf(`
{
  "links": {
    "cloud_controller_v3": {
      "href": "%s/v3",
      "meta": {
        "version": "3.0.0-alpha.5"
      }
    }
  }
}`, server.URL())
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/"),
					VerifyHeaderKV("User-Agent", expectedUserAgent),
					RespondWith(http.StatusOK, rootResponse),
				),
			)

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v3"),
					VerifyHeaderKV("User-Agent", expectedUserAgent),
					RespondWith(http.StatusOK, "{}"),
				),
			)
		})

		It("adds a user agent header", func() {
			_, _, _, err := client.Info()
			Expect(err).ToNot(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(4))
		})
	})
})
