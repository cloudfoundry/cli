package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Actions", func() {
	var (
		actor         *Actor
		fakeUAAClient *v2actionfakes.FakeUAAClient
		fakeConfig    *v2actionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeUAAClient = new(v2actionfakes.FakeUAAClient)
		fakeConfig = new(v2actionfakes.FakeConfig)
		actor = NewActor(nil, fakeUAAClient, fakeConfig)
	})

	Describe("Authenticate", func() {
		var (
			grantType constant.GrantType
			actualErr error
		)

		JustBeforeEach(func() {
			actualErr = actor.Authenticate("some-username", "some-password", "uaa", grantType)
		})

		It("unsets org and space targeting", func() {
			Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(1))
		})

		When("no API errors occur", func() {
			BeforeEach(func() {
				fakeUAAClient.AuthenticateReturns(
					"some-access-token",
					"some-refresh-token",
					nil,
				)
			})

			It("sets the auth and refresh tokens in the config", func() {
				Expect(fakeConfig.SetTokenInformationCallCount()).To(Equal(1))
				accessToken, refreshToken, sshOAuthClient := fakeConfig.SetTokenInformationArgsForCall(0)
				Expect(accessToken).To(Equal("bearer some-access-token"))
				Expect(refreshToken).To(Equal("some-refresh-token"))
				Expect(sshOAuthClient).To(BeEmpty())
			})

			When("the grant type is a password grant", func() {
				BeforeEach(func() {
					grantType = constant.GrantTypePassword
				})

				It("authenticates the user and returns access and refresh tokens", func() {
					Expect(actualErr).NotTo(HaveOccurred())

					Expect(fakeUAAClient.AuthenticateCallCount()).To(Equal(1))
					creds, origin, passedGrantType := fakeUAAClient.AuthenticateArgsForCall(0)
					Expect(creds).To(Equal(map[string]string{
						"username": "some-username",
						"password": "some-password",
					}))
					Expect(origin).To(Equal("uaa"))
					Expect(passedGrantType).To(Equal(constant.GrantTypePassword))

					Expect(fakeConfig.SetUAAGrantTypeCallCount()).To(Equal(1))
					Expect(fakeConfig.SetUAAGrantTypeArgsForCall(0)).To(Equal(""))
				})
			})

			When("the grant type is client credentials", func() {
				BeforeEach(func() {
					grantType = constant.GrantTypeClientCredentials
				})

				It("stores the grant type and the client id", func() {
					Expect(fakeConfig.SetUAAClientCredentialsCallCount()).To(Equal(1))
					clientID, clientSecret := fakeConfig.SetUAAClientCredentialsArgsForCall(0)
					Expect(clientID).To(Equal("some-username"))
					Expect(clientSecret).To(BeEmpty())
					Expect(fakeConfig.SetUAAGrantTypeCallCount()).To(Equal(1))
					Expect(fakeConfig.SetUAAGrantTypeArgsForCall(0)).To(Equal(string(constant.GrantTypeClientCredentials)))
				})

				It("authenticates the user and returns access and refresh tokens", func() {
					Expect(actualErr).NotTo(HaveOccurred())

					Expect(fakeUAAClient.AuthenticateCallCount()).To(Equal(1))
					creds, origin, passedGrantType := fakeUAAClient.AuthenticateArgsForCall(0)
					Expect(creds).To(Equal(map[string]string{
						"client_id":     "some-username",
						"client_secret": "some-password",
					}))
					Expect(origin).To(Equal("uaa"))
					Expect(passedGrantType).To(Equal(constant.GrantTypeClientCredentials))

					Expect(fakeConfig.SetUAAGrantTypeCallCount()).To(Equal(1))
					Expect(fakeConfig.SetUAAGrantTypeArgsForCall(0)).To(BeEquivalentTo(constant.GrantTypeClientCredentials))
				})
			})
		})

		When("an API error occurs", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeUAAClient.AuthenticateReturns(
					"",
					"",
					expectedErr,
				)
			})

			It("returns the error", func() {
				Expect(actualErr).To(MatchError(expectedErr))

				Expect(fakeConfig.SetTokenInformationCallCount()).To(Equal(1))
				accessToken, refreshToken, sshOAuthClient := fakeConfig.SetTokenInformationArgsForCall(0)
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
				Expect(sshOAuthClient).To(BeEmpty())

				Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(1))
			})
		})
	})

	Describe("Revoke", func() {
		var (
			expectedAccessToken  string
			expectedRefreshToken string
		)

		BeforeEach(func() {
			expectedAccessToken = "eyJhbGciOiJSUzI1NiIsImprdSI6Imh0dHBzOi8vdWFhLmdlb2RlLWJhbmUubGl0ZS5jbGkuZnVuL3Rva2VuX2tleXMiLCJraWQiOiJrZXktMSIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIzN2IyZjA0NmIyMGY0MjY1OGY1MWYzZGY4NDZhNjFhOSIsInN1YiI6IjExMjc0OTU2LTYyNTUtNDAyNi05MjUzLWM4ZjE0OWMxZDBkOCIsInNjb3BlIjpbImNsb3VkX2NvbnRyb2xsZXIuYWRtaW4iLCJyb3V0aW5nLnJvdXRlcl9ncm91cHMucmVhZCIsImNsb3VkX2NvbnRyb2xsZXIud3JpdGUiLCJuZXR3b3JrLmFkbWluIiwiZG9wcGxlci5maXJlaG9zZSIsImNsaWVudHMucmVhZCIsInVhYS5yZXNvdXJjZSIsIm9wZW5pZCIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy53cml0ZSIsInNjaW0ucmVhZCIsInVhYS51c2VyIiwiY2xvdWRfY29udHJvbGxlci5yZWFkIiwicGFzc3dvcmQud3JpdGUiLCJzY2ltLndyaXRlIl0sImNsaWVudF9pZCI6ImNmIiwiY2lkIjoiY2YiLCJhenAiOiJjZiIsInJldm9jYWJsZSI6dHJ1ZSwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6IjExMjc0OTU2LTYyNTUtNDAyNi05MjUzLWM4ZjE0OWMxZDBkOCIsIm9yaWdpbiI6InVhYSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsImF1dGhfdGltZSI6MTU5NjY2Nzk0NywicmV2X3NpZyI6ImNlNzJmN2IzIiwiaWF0IjoxNTk2NjY3OTQ3LCJleHAiOjE1OTY2Njg1NDcsImlzcyI6Imh0dHBzOi8vdWFhLmdlb2RlLWJhbmUubGl0ZS5jbGkuZnVuL29hdXRoL3Rva2VuIiwiemlkIjoidWFhIiwiYXVkIjpbImNsb3VkX2NvbnRyb2xsZXIiLCJzY2ltIiwicGFzc3dvcmQiLCJjZiIsImNsaWVudHMiLCJ1YWEiLCJvcGVuaWQiLCJkb3BwbGVyIiwicm91dGluZy5yb3V0ZXJfZ3JvdXBzIiwibmV0d29yayJdfQ.WGRIexU6aheL6VOxK3lOCdQzv8GNK_O1IHu77pyEl0Y5wxa10T1H0VppzpKPbxYJ1C9QvQp1UwTk8fU9ylJ6lFAhA1xXLpBHNMI66QSVYc2m2TmOMUbJjWoE3w7cF-cBM1Unle3JlRk7y8yK4jeo0Gwj1fWcE4JNDKypTrT7yd5WxRCcMk1pWTJtLH9Rgng0Ei6i15NJ9K2SbF-rQWQp0qr1l4PBcyrjf7hWbf365yZFUSDMtPAToiJeKT3qmmN37elDmvFYzFNN4MEqEoirWYfjjakulLh9TPWELiygwNzDi11MYO07ksCUow9ArcA7cSlpnh9qrfPOqlfZTizBGg"
			expectedRefreshToken = "eyJhbGciOiJSUzI1NiIsImprdSI6Imh0dHBzOi8vdWFhLmdlb2RlLWJhbmUubGl0ZS5jbGkuZnVuL3Rva2VuX2tleXMiLCJraWQiOiJrZXktMSIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIzN2IyZjA0NmIyMGY0MjY1OGY1MWYzZGY4NDZhNjFhOSIsInN1YiI6IjExMjc0OTU2LTYyNTUtNDAyNi05MjUzLWM4ZjE0OWMxZDBkOCIsInNjb3BlIjpbImNsb3VkX2NvbnRyb2xsZXIuYWRtaW4iLCJyb3V0aW5nLnJvdXRlcl9ncm91cHMucmVhZCIsImNsb3VkX2NvbnRyb2xsZXIud3JpdGUiLCJuZXR3b3JrLmFkbWluIiwiZG9wcGxlci5maXJlaG9zZSIsImNsaWVudHMucmVhZCIsInVhYS5yZXNvdXJjZSIsIm9wZW5pZCIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy53cml0ZSIsInNjaW0ucmVhZCIsInVhYS51c2VyIiwiY2xvdWRfY29udHJvbGxlci5yZWFkIiwicGFzc3dvcmQud3JpdGUiLCJzY2ltLndyaXRlIl0sImNsaWVudF9pZCI6ImNmIiwiY2lkIjoiY2YiLCJhenAiOiJjZiIsInJldm9jYWJsZSI6dHJ1ZSwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6IjExMjc0OTU2LTYyNTUtNDAyNi05MjUzLWM4ZjE0OWMxZDBkOCIsIm9yaWdpbiI6InVhYSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsImF1dGhfdGltZSI6MTU5NjY2Nzk0NywicmV2X3NpZyI6ImNlNzJmN2IzIiwiaWF0IjoxNTk2NjY3OTQ3LCJleHAiOjE1OTY2Njg1NDcsImlzcyI6Imh0dHBzOi8vdWFhLmdlb2RlLWJhbmUubGl0ZS5jbGkuZnVuL29hdXRoL3Rva2VuIiwiemlkIjoidWFhIiwiYXVkIjpbImNsb3VkX2NvbnRyb2xsZXIiLCJzY2ltIiwicGFzc3dvcmQiLCJjZiIsImNsaWVudHMiLCJ1YWEiLCJvcGVuaWQiLCJkb3BwbGVyIiwicm91dGluZy5yb3V0ZXJfZ3JvdXBzIiwibmV0d29yayJdfQ.WGRIexU6aheL6VOxK3lOCdQzv8GNK_O1IHu77pyEl0Y5wxa10T1H0VppzpKPbxYJ1C9QvQp1UwTk8fU9ylJ6lFAhA1xXLpBHNMI66QSVYc2m2TmOMUbJjWoE3w7cF-cBM1Unle3JlRk7y8yK4jeo0Gwj1fWcE4JNDKypTrT7yd5WxRCcMk1pWTJtLH9Rgng0Ei6i15NJ9K2SbF-rQWQp0qr1l4PBcyrjf7hWbf365yZFUSDMtPAToiJeKT3qmmN37elDmvFYzFNN4MEqEoirWYfjjakulLh9TPWELiygwNzDi11MYO07ksCUow9ArcA7cSlpnh9qrfPOqlfZTizBGg"
			fakeConfig.AccessTokenReturns(expectedAccessToken)
			fakeConfig.RefreshTokenReturns(expectedRefreshToken)
		})

		JustBeforeEach(func() {
			_ = actor.Revoke()
		})

		When("the token is revokable", func() {

			It("calls the UAA to revoke refresh and access tokens", func() {
				Expect(fakeUAAClient.RevokeCallCount()).To(Equal(2))

				Expect(fakeUAAClient.RevokeArgsForCall(0)).To(Equal(expectedRefreshToken))
				Expect(fakeUAAClient.RevokeArgsForCall(1)).To(Equal(expectedAccessToken))
			})
		})

		When("the token is not revokable", func() {
			BeforeEach(func() {
				expectedAccessToken = "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMTI3NDk1Ni02MjU1LTQwMjYtOTI1My1jOGYxNDljMWQwZDgiLCJhdWQiOltdfQ.rVlUqVTglhod_MEbdnczGwj4IJIMHqiLrqaX2wvEWMw"
				fakeConfig.AccessTokenReturns(expectedAccessToken)
			})

			It("does not call the UAA to revoke refresh and access tokens", func() {
				Expect(fakeUAAClient.RevokeCallCount()).To(Equal(0))
			})
		})
	})

	Describe("GetLoginPrompts", func() {
		When("getting login prompts info from UAA", func() {
			var (
				prompts map[string]coreconfig.AuthPrompt
			)

			BeforeEach(func() {
				fakeUAAClient.LoginPromptsReturns(map[string][]string{
					"username": {"text", "Email"},
					"pin":      {"password", "PIN Number"},
				})
				prompts = actor.GetLoginPrompts()
			})

			It("gets the login prompts", func() {
				Expect(prompts).To(Equal(map[string]coreconfig.AuthPrompt{
					"username": {
						DisplayName: "Email",
						Type:        coreconfig.AuthPromptTypeText,
					},
					"pin": {
						DisplayName: "PIN Number",
						Type:        coreconfig.AuthPromptTypePassword,
					},
				}))
			})
		})
	})
})
