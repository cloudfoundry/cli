package configaction_test

import (
	. "code.cloudfoundry.org/cli/actor/configaction"
	"code.cloudfoundry.org/cli/actor/configaction/configactionfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Targgeting", func() {
	var (
		actor             Actor
		skipSSLValidation bool

		fakeCloudControllerClient *configactionfakes.FakeCloudControllerClient
		fakeConfig                *configactionfakes.FakeConfig
		settings                  TargetSettings
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(configactionfakes.FakeCloudControllerClient)
		fakeConfig = new(configactionfakes.FakeConfig)
		actor = NewActor(fakeConfig, fakeCloudControllerClient)

		settings = TargetSettings{
			SkipSSLValidation: skipSSLValidation,
		}
	})

	Describe("SetTarget", func() {
		var expectedAPI, expectedAPIVersion, expectedAuth, expectedLoggregator, expectedDoppler, expectedUAA, expectedRouting string

		BeforeEach(func() {
			expectedAPI = "https://api.foo.com"
			expectedAPIVersion = "2.59.0"
			expectedAuth = "https://login.foo.com"
			expectedLoggregator = "wss://log.foo.com"
			expectedDoppler = "wss://doppler.foo.com"
			expectedUAA = "https://uaa.foo.com"
			expectedRouting = "https://api.foo.com/routing"

			settings.URL = expectedAPI

			fakeCloudControllerClient.APIReturns(expectedAPI)
			fakeCloudControllerClient.APIVersionReturns(expectedAPIVersion)
			fakeCloudControllerClient.AuthorizationEndpointReturns(expectedAuth)
			fakeCloudControllerClient.LoggregatorEndpointReturns(expectedLoggregator)
			fakeCloudControllerClient.DopplerEndpointReturns(expectedDoppler)
			fakeCloudControllerClient.TokenEndpointReturns(expectedUAA)
			fakeCloudControllerClient.RoutingEndpointReturns(expectedRouting)
		})

		It("targets the passed API", func() {
			_, err := actor.SetTarget(settings)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeCloudControllerClient.TargetCFCallCount()).To(Equal(1))
			connectionSettings := fakeCloudControllerClient.TargetCFArgsForCall(0)
			Expect(connectionSettings.URL).To(Equal(expectedAPI))
			Expect(connectionSettings.SkipSSLValidation).To(BeFalse())
		})

		It("sets all the target information", func() {
			_, err := actor.SetTarget(settings)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeConfig.SetTargetInformationCallCount()).To(Equal(1))
			api, apiVersion, auth, loggregator, doppler, uaa, routing, sslDisabled := fakeConfig.SetTargetInformationArgsForCall(0)

			Expect(api).To(Equal(expectedAPI))
			Expect(apiVersion).To(Equal(expectedAPIVersion))
			Expect(auth).To(Equal(expectedAuth))
			Expect(loggregator).To(Equal(expectedLoggregator))
			Expect(doppler).To(Equal(expectedDoppler))
			Expect(uaa).To(Equal(expectedUAA))
			Expect(routing).To(Equal(expectedRouting))
			Expect(sslDisabled).To(Equal(skipSSLValidation))
		})

		It("clears all the token information", func() {
			_, err := actor.SetTarget(settings)
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
				warnings, err := actor.SetTarget(settings)
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(BeNil())

				Expect(fakeCloudControllerClient.TargetCFCallCount()).To(BeZero())
			})
		})
	})

	Describe("ClearTarget", func() {
		It("clears all the target information", func() {
			actor.ClearTarget()

			Expect(fakeConfig.SetTargetInformationCallCount()).To(Equal(1))
			api, apiVersion, auth, loggregator, doppler, uaa, routing, sslDisabled := fakeConfig.SetTargetInformationArgsForCall(0)

			Expect(api).To(BeEmpty())
			Expect(apiVersion).To(BeEmpty())
			Expect(auth).To(BeEmpty())
			Expect(loggregator).To(BeEmpty())
			Expect(doppler).To(BeEmpty())
			Expect(uaa).To(BeEmpty())
			Expect(routing).To(BeEmpty())
			Expect(sslDisabled).To(BeFalse())
		})

		It("clears all the token information", func() {
			actor.ClearTarget()

			Expect(fakeConfig.SetTokenInformationCallCount()).To(Equal(1))
			accessToken, refreshToken, sshOAuthClient := fakeConfig.SetTokenInformationArgsForCall(0)

			Expect(accessToken).To(BeEmpty())
			Expect(refreshToken).To(BeEmpty())
			Expect(sshOAuthClient).To(BeEmpty())
		})
	})
})
