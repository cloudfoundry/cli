package plugin_test

import (
	"fmt"
	"net/http"
	"net/url"

	. "code.cloudfoundry.org/cli/api/plugin"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("PluginRepository", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetPluginRepository", func() {
		Context("when the url points to a valid CF CLI plugin repo", func() {
			var response string

			BeforeEach(func() {
				response = `{
					"plugins": [
						{
							"name": "plugin-1",
							"description": "useful plugin for useful things",
							"version": "1.0.0",
							"binaries": [{"platform":"osx","url":"http://some-url","checksum":"somechecksum"},{"platform":"win64","url":"http://another-url","checksum":"anotherchecksum"},{"platform":"linux64","url":"http://last-url","checksum":"lastchecksum"}]
						},
						{
							"name": "plugin-2",
							"description": "amazing plugin",
							"version": "1.0.0"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/list"),
						RespondWith(http.StatusOK, response),
					),
				)
			})

			It("returns the plugin repository", func() {
				pluginRepository, err := client.GetPluginRepository(server.URL())
				Expect(err).ToNot(HaveOccurred())
				Expect(pluginRepository).To(Equal(PluginRepository{
					Plugins: []Plugin{
						{
							Name:        "plugin-1",
							Description: "useful plugin for useful things",
							Version:     "1.0.0",
							Binaries: []PluginBinary{
								{Platform: "osx", URL: "http://some-url", Checksum: "somechecksum"},
								{Platform: "win64", URL: "http://another-url", Checksum: "anotherchecksum"},
								{Platform: "linux64", URL: "http://last-url", Checksum: "lastchecksum"},
							},
						},
						{
							Name:        "plugin-2",
							Description: "amazing plugin",
							Version:     "1.0.0",
						},
					},
				}))
			})

			Context("when the URL has a trailing slash", func() {
				It("still hits the /list endpoint on the URL", func() {
					_, err := client.GetPluginRepository(fmt.Sprintf("%s/", server.URL()))
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("when the URL has a trailing '/list'", func() {
				It("still hits the /list endpoint on the URL", func() {
					_, err := client.GetPluginRepository(fmt.Sprintf("%s/list", server.URL()))
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("when the URL has a trailing '/list/'", func() {
				It("still hits the /list endpoint on the URL", func() {
					_, err := client.GetPluginRepository(fmt.Sprintf("%s/list/", server.URL()))
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("when the URL has path different from /list", func() {
				BeforeEach(func() {
					server.SetHandler(0,
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/cli/list"),
							RespondWith(http.StatusOK, response),
						),
					)
				})

				It("appends /list to the path", func() {
					_, err := client.GetPluginRepository(fmt.Sprintf("%s/cli", server.URL()))
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when the repository URL in invalid", func() {
			It("returns an error", func() {
				_, err := client.GetPluginRepository("http://not a valid URL")
				Expect(err).To(BeAssignableToTypeOf(&url.Error{}))
			})
		})

		Context("when the server returns an error", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/list"),
						RespondWith(http.StatusNotFound, nil),
					),
				)
			})

			It("returns the error", func() {
				_, err := client.GetPluginRepository(server.URL())
				Expect(err).To(MatchError(pluginerror.RawHTTPStatusError{Status: "404 Not Found", RawResponse: []byte{}}))
			})
		})
	})
})
