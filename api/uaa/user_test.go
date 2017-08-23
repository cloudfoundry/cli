package uaa_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("User", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestUAAClientAndStore()
	})

	Describe("CreateUser", func() {
		Context("when no errors occur", func() {
			Context("when creating user with origin", func() {
				BeforeEach(func() {
					response := `{
					"ID": "new-user-id"
				}`
					uaaServer.AppendHandlers(
						CombineHandlers(
							verifyRequestHost(TestUAAResource),
							VerifyRequest(http.MethodPost, "/Users"),
							VerifyHeaderKV("Content-Type", "application/json"),
							VerifyBody([]byte(`{"userName":"new-user","password":"","origin":"some-origin","name":{"familyName":"new-user","givenName":"new-user"},"emails":[{"value":"new-user","primary":true}]}`)),
							RespondWith(http.StatusOK, response),
						))
				})

				It("creates a new user", func() {
					user, err := client.CreateUser("new-user", "", "some-origin")
					Expect(err).NotTo(HaveOccurred())

					Expect(user).To(Equal(User{
						ID: "new-user-id",
					}))
				})
			})
			Context("when creating user in UAA", func() {
				BeforeEach(func() {
					response := `{
					"ID": "new-user-id"
				}`
					uaaServer.AppendHandlers(
						CombineHandlers(
							verifyRequestHost(TestUAAResource),
							VerifyRequest(http.MethodPost, "/Users"),
							VerifyHeaderKV("Content-Type", "application/json"),
							VerifyBody([]byte(`{"userName":"new-user","password":"new-password","origin":"","name":{"familyName":"new-user","givenName":"new-user"},"emails":[{"value":"new-user","primary":true}]}`)),
							RespondWith(http.StatusOK, response),
						))
				})

				It("creates a new user", func() {
					user, err := client.CreateUser("new-user", "new-password", "")
					Expect(err).NotTo(HaveOccurred())

					Expect(user).To(Equal(User{
						ID: "new-user-id",
					}))
				})
			})
		})

		Context("when an error occurs", func() {
			var response string

			BeforeEach(func() {
				response = `{
					"error": "some-error",
					"error_description": "some-description"
				}`
				uaaServer.AppendHandlers(
					CombineHandlers(
						verifyRequestHost(TestUAAResource),
						VerifyRequest(http.MethodPost, "/Users"),
						RespondWith(http.StatusTeapot, response),
					))
			})

			It("returns the error", func() {
				_, err := client.CreateUser("new-user", "new-password", "")
				Expect(err).To(MatchError(RawHTTPStatusError{
					StatusCode:  http.StatusTeapot,
					RawResponse: []byte(response),
				}))
			})
		})
	})
})
