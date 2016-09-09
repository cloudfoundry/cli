package configactions_test

import (
	. "code.cloudfoundry.org/cli/actors/configactions"
	"code.cloudfoundry.org/cli/actors/configactions/configactionsfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Targgeting", func() {
	var (
		actor             Actor
		skipSSLValidation bool

		fakeCloudControllerClient *configactionsfakes.FakeCloudControllerClient
		fakeConfig                *configactionsfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(configactionsfakes.FakeCloudControllerClient)
		fakeConfig = new(configactionsfakes.FakeConfig)
		actor = NewActor(fakeConfig, fakeCloudControllerClient)
	})

	Describe("SetTarget", func() {
		var expectedAPI, expectedAPIVersion, expectedAuth, expectedLoggregator, expectedDoppler, expectedUAA string

		BeforeEach(func() {
			expectedAPI = "https://api.foo.com"
			expectedAPIVersion = "2.59.0"
			expectedAuth = "https://login.foo.com"
			expectedLoggregator = "wss://log.foo.com"
			expectedDoppler = "wss://doppler.foo.com"
			expectedUAA = "https://uaa.foo.com"

			fakeCloudControllerClient.APIReturns(expectedAPI)
			fakeCloudControllerClient.APIVersionReturns(expectedAPIVersion)
			fakeCloudControllerClient.AuthorizationEndpointReturns(expectedAuth)
			fakeCloudControllerClient.LoggregatorEndpointReturns(expectedLoggregator)
			fakeCloudControllerClient.DopplerEndpointReturns(expectedDoppler)
			fakeCloudControllerClient.TokenEndpointReturns(expectedUAA)
		})

		It("targets the passed API", func() {
			_, err := actor.SetTarget(expectedAPI, skipSSLValidation)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeCloudControllerClient.TargetCFCallCount()).To(Equal(1))
			api, skipSSL := fakeCloudControllerClient.TargetCFArgsForCall(0)
			Expect(api).To(Equal(expectedAPI))
			Expect(skipSSL).To(BeFalse())
		})

		It("sets all the target information", func() {
			_, err := actor.SetTarget(expectedAPI, skipSSLValidation)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeConfig.SetTargetInformationCallCount()).To(Equal(1))
			api, apiVersion, auth, loggregator, doppler, uaa, sslDisabled := fakeConfig.SetTargetInformationArgsForCall(0)

			Expect(api).To(Equal(expectedAPI))
			Expect(apiVersion).To(Equal(expectedAPIVersion))
			Expect(auth).To(Equal(expectedAuth))
			Expect(loggregator).To(Equal(expectedLoggregator))
			Expect(doppler).To(Equal(expectedDoppler))
			Expect(uaa).To(Equal(expectedUAA))
			Expect(sslDisabled).To(Equal(skipSSLValidation))
		})
	})

	Describe("ClearTarget", func() {
		It("clears all the target information", func() {
			actor.ClearTarget()

			Expect(fakeConfig.SetTargetInformationCallCount()).To(Equal(1))
			api, apiVersion, auth, loggregator, doppler, uaa, sslDisabled := fakeConfig.SetTargetInformationArgsForCall(0)

			Expect(api).To(BeEmpty())
			Expect(apiVersion).To(BeEmpty())
			Expect(auth).To(BeEmpty())
			Expect(loggregator).To(BeEmpty())
			Expect(doppler).To(BeEmpty())
			Expect(uaa).To(BeEmpty())
			Expect(sslDisabled).To(BeFalse())
		})
	})
})
