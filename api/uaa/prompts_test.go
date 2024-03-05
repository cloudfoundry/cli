package uaa_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/uaa"
	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Prompts", func() {
	var (
		client     *Client
		fakeConfig *uaafakes.FakeConfig
	)

	BeforeEach(func() {
		fakeConfig = NewTestConfig()
		client = NewClient(fakeConfig)

		client.Info.Links.Login = "https://" + TestAuthorizationResource
	})

	Describe("GetLoginPrompts", func() {
		var (
			prompts map[string][]string
			err     error
		)

		JustBeforeEach(func() {
			prompts, err = client.GetLoginPrompts()
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				response := `{
					"prompts": {
						"one": ["a", "b"],
						"two": ["c", "d"]
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						verifyRequestHost(TestAuthorizationResource),
						VerifyRequest(http.MethodGet, "/login"),
						RespondWith(http.StatusOK, response),
					),
				)
			})

			It("returns the login prompts", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(prompts).To(Equal(map[string][]string{
					"one": {"a", "b"},
					"two": {"c", "d"},
				}))
			})
		})

		When("the request fails", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						verifyRequestHost(TestAuthorizationResource),
						VerifyRequest(http.MethodGet, "/login"),
						RespondWith(http.StatusTeapot, `{}`),
					),
				)
			})

			It("returns the error", func() {
				Expect(err).To(MatchError(uaa.RawHTTPStatusError{
					StatusCode:  http.StatusTeapot,
					RawResponse: []byte(`{}`),
				}))
			})
		})
	})
})
