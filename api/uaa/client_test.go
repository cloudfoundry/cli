package uaa_test

import (
	"fmt"
	"net/http"
	"runtime"

	. "code.cloudfoundry.org/cli/api/uaa"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("UAA Client", func() {
	var (
		client *Client
	)

	BeforeEach(func() {
		client = NewTestUAAClientAndStore()
	})

	Describe("Request Headers", func() {
		Describe("User-Agent", func() {
			var userAgent string
			BeforeEach(func() {
				userAgent = fmt.Sprintf("CF CLI UAA API Test/Unknown (%s; %s %s)",
					runtime.Version(),
					runtime.GOARCH,
					runtime.GOOS,
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/oauth/token"),
						VerifyHeaderKV("User-Agent", userAgent),
						RespondWith(http.StatusOK, "{}"),
					))
			})

			It("adds the User-Agent header to requests", func() {
				_, err := client.RefreshAccessToken("")
				Expect(err).ToNot(HaveOccurred())

				Expect(server.ReceivedRequests()).To(HaveLen(2))
			})
		})

		Describe("Conection", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/oauth/token"),
						VerifyHeaderKV("Connection", "close"),
						RespondWith(http.StatusOK, "{}"),
					))
			})

			It("forcefully closes the connection after each request", func() {
				_, err := client.RefreshAccessToken("")
				Expect(err).ToNot(HaveOccurred())

				Expect(server.ReceivedRequests()).To(HaveLen(2))
			})
		})
	})
})
