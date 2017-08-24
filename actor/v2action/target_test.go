package v2action_test

import (
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Targeting", func() {
	var (
		actor             *Actor
		skipSSLValidation bool

		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		fakeConfig                *v2actionfakes.FakeConfig
		settings                  TargetSettings
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v2actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, nil, nil)

		settings = TargetSettings{
			SkipSSLValidation: skipSSLValidation,
		}
	})

	Describe("SetTarget", func() {
		var expectedAPI, expectedAPIVersion, expectedAuth, expectedMinCLIVersion, expectedDoppler, expectedRouting string

		BeforeEach(func() {
			expectedAPI = "https://api.foo.com"
			expectedAPIVersion = "2.59.0"
			expectedAuth = "https://login.foo.com"
			expectedMinCLIVersion = "1.0.0"
			expectedDoppler = "wss://doppler.foo.com"
			expectedRouting = "https://api.foo.com/routing"

			settings.URL = expectedAPI

			fakeCloudControllerClient.APIReturns(expectedAPI)
			fakeCloudControllerClient.APIVersionReturns(expectedAPIVersion)
			fakeCloudControllerClient.AuthorizationEndpointReturns(expectedAuth)
			fakeCloudControllerClient.MinCLIVersionReturns(expectedMinCLIVersion)
			fakeCloudControllerClient.DopplerEndpointReturns(expectedDoppler)
			fakeCloudControllerClient.RoutingEndpointReturns(expectedRouting)
		})

		It("targets the passed API", func() {
			_, err := actor.SetTarget(fakeConfig, settings)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeCloudControllerClient.TargetCFCallCount()).To(Equal(1))
			connectionSettings := fakeCloudControllerClient.TargetCFArgsForCall(0)
			Expect(connectionSettings.URL).To(Equal(expectedAPI))
			Expect(connectionSettings.SkipSSLValidation).To(BeFalse())
		})

		It("sets all the target information", func() {
			_, err := actor.SetTarget(fakeConfig, settings)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeConfig.SetTargetInformationCallCount()).To(Equal(1))
			api, apiVersion, auth, minCLIVersion, doppler, routing, sslDisabled := fakeConfig.SetTargetInformationArgsForCall(0)

			Expect(api).To(Equal(expectedAPI))
			Expect(apiVersion).To(Equal(expectedAPIVersion))
			Expect(auth).To(Equal(expectedAuth))
			Expect(minCLIVersion).To(Equal(expectedMinCLIVersion))
			Expect(doppler).To(Equal(expectedDoppler))
			Expect(routing).To(Equal(expectedRouting))
			Expect(sslDisabled).To(Equal(skipSSLValidation))
		})

		It("clears all the token information", func() {
			_, err := actor.SetTarget(fakeConfig, settings)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeConfig.SetTokenInformationCallCount()).To(Equal(1))
			accessToken, refreshToken, sshOAuthClient := fakeConfig.SetTokenInformationArgsForCall(0)

			Expect(accessToken).To(BeEmpty())
			Expect(refreshToken).To(BeEmpty())
			Expect(sshOAuthClient).To(BeEmpty())
		})

		Context("when setting the same API and skip SSL configuration", func() {
			var APIURL string

			BeforeEach(func() {
				APIURL = "https://some-api.com"
				settings.URL = APIURL
				fakeConfig.TargetReturns(APIURL)
				fakeConfig.SkipSSLValidationReturns(skipSSLValidation)
			})

			It("does not make any API calls", func() {
				warnings, err := actor.SetTarget(fakeConfig, settings)
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(BeNil())

				Expect(fakeCloudControllerClient.TargetCFCallCount()).To(BeZero())
			})
		})
	})

	Describe("ClearTarget", func() {
		It("clears all the target information", func() {
			actor.ClearTarget(fakeConfig)

			Expect(fakeConfig.SetTargetInformationCallCount()).To(Equal(1))
			api, apiVersion, auth, minCLIVersion, doppler, routing, sslDisabled := fakeConfig.SetTargetInformationArgsForCall(0)

			Expect(api).To(BeEmpty())
			Expect(apiVersion).To(BeEmpty())
			Expect(auth).To(BeEmpty())
			Expect(minCLIVersion).To(BeEmpty())
			Expect(doppler).To(BeEmpty())
			Expect(routing).To(BeEmpty())
			Expect(sslDisabled).To(BeFalse())
		})

		It("clears all the token information", func() {
			actor.ClearTarget(fakeConfig)

			Expect(fakeConfig.SetTokenInformationCallCount()).To(Equal(1))
			accessToken, refreshToken, sshOAuthClient := fakeConfig.SetTokenInformationArgsForCall(0)

			Expect(accessToken).To(BeEmpty())
			Expect(refreshToken).To(BeEmpty())
			Expect(sshOAuthClient).To(BeEmpty())
		})
	})

	Describe("ClearOrganizationAndSpace", func() {
		It("clears all organization and space information", func() {
			actor.ClearOrganizationAndSpace(fakeConfig)

			Expect(fakeConfig.UnsetOrganizationInformationCallCount()).To(Equal(1))
			Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(1))
		})
	})
})
