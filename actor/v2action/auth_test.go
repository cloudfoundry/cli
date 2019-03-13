package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
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

		When("no API errors occur", func() {
			BeforeEach(func() {
				fakeUAAClient.AuthenticateReturns(
					"some-access-token",
					"some-refresh-token",
					nil,
				)
			})

			When("the grant type is a password grant", func() {
				BeforeEach(func() {
					grantType = constant.GrantTypePassword
				})

				It("authenticates the user and returns access and refresh tokens", func() {
					Expect(actualErr).NotTo(HaveOccurred())

					Expect(fakeUAAClient.AuthenticateCallCount()).To(Equal(1))
					ID, secret, origin, passedGrantType := fakeUAAClient.AuthenticateArgsForCall(0)
					Expect(ID).To(Equal("some-username"))
					Expect(secret).To(Equal("some-password"))
					Expect(origin).To(Equal("uaa"))
					Expect(passedGrantType).To(Equal(constant.GrantTypePassword))

					Expect(fakeConfig.SetTokenInformationCallCount()).To(Equal(1))
					accessToken, refreshToken, sshOAuthClient := fakeConfig.SetTokenInformationArgsForCall(0)
					Expect(accessToken).To(Equal("bearer some-access-token"))
					Expect(refreshToken).To(Equal("some-refresh-token"))
					Expect(sshOAuthClient).To(BeEmpty())

					Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(1))
					Expect(fakeConfig.SetUAAGrantTypeCallCount()).To(Equal(0))
				})

				When("a previous user authenticated with a client grant type", func() {
					BeforeEach(func() {
						fakeConfig.UAAGrantTypeReturns("client_credentials")
					})
					It("returns a PasswordGrantTypeLogoutRequiredError", func() {
						Expect(actualErr).To(MatchError(actionerror.PasswordGrantTypeLogoutRequiredError{}))
						Expect(fakeConfig.UAAGrantTypeCallCount()).To(Equal(1))
					})
				})
			})

			When("the grant type is not password", func() {
				BeforeEach(func() {
					grantType = constant.GrantTypeClientCredentials
				})

				It("stores the grant type and the client credentials", func() {
					Expect(fakeConfig.SetUAAClientCredentialsCallCount()).To(Equal(1))
					client, clientSecret := fakeConfig.SetUAAClientCredentialsArgsForCall(0)
					Expect(client).To(Equal("some-username"))
					Expect(clientSecret).To(Equal("some-password"))
					Expect(fakeConfig.SetUAAGrantTypeCallCount()).To(Equal(1))
					Expect(fakeConfig.SetUAAGrantTypeArgsForCall(0)).To(Equal(string(constant.GrantTypeClientCredentials)))
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
