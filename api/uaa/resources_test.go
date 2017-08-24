package uaa_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("SetupResources", func() {
	var (
		client            *Client
		setupResourcesErr error
		fakeStore         *uaafakes.FakeUAAEndpointStore
	)

	JustBeforeEach(func() {
		fakeStore = new(uaafakes.FakeUAAEndpointStore)
		setupResourcesErr = client.SetupResources(fakeStore, server.URL())
	})

	BeforeEach(func() {
		client = NewClient(Config{
			AppName:           "CF CLI UAA API Test",
			AppVersion:        "Unknown",
			ClientID:          "client-id",
			ClientSecret:      "client-secret",
			SkipSSLValidation: true,
		})
	})

	Context("when the authentication server returns an error", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/login"),
					RespondWith(http.StatusNotFound, `{"errors": [{}]}`, nil),
				),
			)
		})

		It("returns the error", func() {
			Expect(setupResourcesErr).To(HaveOccurred())
			Expect(fakeStore.SetUAAEndpointCallCount()).To(Equal(0))
		})
	})

	Context("when the request succeeds", func() {
		Context("and the UAA field is populated", func() {
			BeforeEach(func() {
				response := `{
				"links": {
					"uaa": "https://uaa.bosh-lite.com"
				}
			}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/login"),
						RespondWith(http.StatusOK, response, nil),
					),
				)
			})

			It("sets the UAA endpoint to the UAA link and does not return an error", func() {
				Expect(setupResourcesErr).ToNot(HaveOccurred())
				Expect(fakeStore.SetUAAEndpointCallCount()).To(Equal(1))
				Expect(fakeStore.SetUAAEndpointArgsForCall(0)).To(Equal("https://uaa.bosh-lite.com"))
			})
		})

		Context("when the UAA field is not populated", func() {
			BeforeEach(func() {
				response := `{
				"links": {}
			}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/login"),
						RespondWith(http.StatusOK, response, nil),
					),
				)
			})

			It("sets the UAA endpoint to the bootstrap endpoint and does not return an error", func() {
				Expect(setupResourcesErr).ToNot(HaveOccurred())
				Expect(fakeStore.SetUAAEndpointCallCount()).To(Equal(1))
				Expect(fakeStore.SetUAAEndpointArgsForCall(0)).To(Equal(server.URL()))
			})
		})
	})
})
