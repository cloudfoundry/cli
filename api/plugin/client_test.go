package plugin_test

import (
	"fmt"
	"net/http"
	"runtime"

	. "code.cloudfoundry.org/cli/api/plugin"
	"code.cloudfoundry.org/cli/api/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Plugin Client", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("WrapConnection", func() {
		var fakeConnectionWrapper *pluginfakes.FakeConnectionWrapper

		BeforeEach(func() {
			fakeConnectionWrapper = new(pluginfakes.FakeConnectionWrapper)
			fakeConnectionWrapper.WrapReturns(fakeConnectionWrapper)
		})

		It("wraps the existing connection in the provided wrapper", func() {
			client.WrapConnection(fakeConnectionWrapper)
			Expect(fakeConnectionWrapper.WrapCallCount()).To(Equal(1))

			client.GetPluginRepository("does-not-matter")
			Expect(fakeConnectionWrapper.MakeCallCount()).To(Equal(1))
		})
	})

	Describe("User Agent", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/list"),
					VerifyHeaderKV("User-Agent", fmt.Sprintf("CF CLI API Plugin Test/Unknown (%s; %s %s)", runtime.Version(), runtime.GOARCH, runtime.GOOS)),
					RespondWith(http.StatusOK, "{}"),
				),
			)
		})

		It("adds a user agent header", func() {
			client.GetPluginRepository(server.URL())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})
	})
})
