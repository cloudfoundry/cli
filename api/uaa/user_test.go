package uaa_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("User", func() {
	var (
		client *Client

		fakeConfig *uaafakes.FakeConfig
	)

	BeforeEach(func() {
		fakeConfig = NewTestConfig()

		client = NewTestUAAClientAndStore(fakeConfig)
	})

	Describe("CreateUser", func() {
		When("no errors occur", func() {
			When("creating user with origin", func() {
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
			When("creating user in UAA", func() {
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

		When("an error occurs", func() {
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

	Describe("GetUsers", func() {
		var (
			userName string
			origin   string
			users    []User
			err      error
		)

		BeforeEach(func() {
			userName = ""
			origin = ""
			users = []User{}
			err = nil
		})

		JustBeforeEach(func() {
			users, err = client.GetUsers(userName, origin)
		})

		When("no errors occur", func() {
			When("getting the users by username", func() {
				BeforeEach(func() {
					userName = "existing-user"
					origin = ""

					response := `{
						"resources": [
							{ "ID": "existing-user-id" }
						]
					}`

					uaaServer.AppendHandlers(
						CombineHandlers(
							verifyRequestHost(TestUAAResource),
							VerifyRequest(http.MethodGet, "/Users", "filter=userName+eq+%22existing-user%22"),
							VerifyHeaderKV("Content-Type", "application/json"),
							RespondWith(http.StatusOK, response),
						))
				})

				It("gets users by username", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(users).To(Equal([]User{
						{ID: "existing-user-id"},
					}))
				})
			})

			When("getting the user by username and origin", func() {
				BeforeEach(func() {
					userName = "existing-user"
					origin = "ldap"

					response := `{
						"resources": [
							{ "ID": "existing-user-id" }
						]
					}`

					uaaServer.AppendHandlers(
						CombineHandlers(
							verifyRequestHost(TestUAAResource),
							VerifyRequest(http.MethodGet, "/Users", "filter=userName+eq+%22existing-user%22+and+origin+eq+%22ldap%22"),
							VerifyHeaderKV("Content-Type", "application/json"),
							RespondWith(http.StatusOK, response),
						))
				})

				It("gets user by username and origin", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(users).To(Equal([]User{
						{ID: "existing-user-id"},
					}))
				})
			})
		})

		When("an error occurs", func() {
			var response string

			BeforeEach(func() {
				userName = "existing-user"
				origin = "ldap"

				response = `{
					"error_description": "Invalid filter expression"
				}`

				uaaServer.AppendHandlers(
					CombineHandlers(
						verifyRequestHost(TestUAAResource),
						VerifyRequest(http.MethodGet, "/Users", "filter=userName+eq+%22existing-user%22+and+origin+eq+%22ldap%22"),
						VerifyHeaderKV("Content-Type", "application/json"),
						RespondWith(http.StatusBadRequest, response),
					))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError(RawHTTPStatusError{
					StatusCode:  400,
					RawResponse: []byte(response),
				}))
			})
		})
	})
})
