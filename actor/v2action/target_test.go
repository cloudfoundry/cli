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
		actor = NewActor(fakeCloudControllerClient, nil, fakeConfig)

		settings = TargetSettings{
			SkipSSLValidation: skipSSLValidation,
		}
	})

	Describe("ClearTarget", func() {
		It("clears all the target information", func() {
			actor.ClearTarget()
			Expect(fakeConfig.SetTargetInformationCallCount()).To(Equal(1))
			targetInfoArgs := fakeConfig.SetTargetInformationArgsForCall(0)

			Expect(targetInfoArgs.Api).To(BeEmpty())
			Expect(targetInfoArgs.ApiVersion).To(BeEmpty())
			Expect(targetInfoArgs.Auth).To(BeEmpty())
			Expect(targetInfoArgs.MinCLIVersion).To(BeEmpty())
			Expect(targetInfoArgs.Doppler).To(BeEmpty())
			Expect(targetInfoArgs.Routing).To(BeEmpty())
			Expect(targetInfoArgs.SkipSSLValidation).To(BeFalse())
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

	Describe("MinCLIVersion", func() {
		var version string

		BeforeEach(func() {
			fakeCloudControllerClient.MinCLIVersionReturns("10.9.8")
		})

		JustBeforeEach(func() {
			version = actor.MinCLIVersion()
		})

		It("returns the version that the API reports", func() {
			Expect(version).To(Equal("10.9.8"))
		})
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
			targetInfoArgs := fakeConfig.SetTargetInformationArgsForCall(0)

			Expect(targetInfoArgs.Api).To(Equal(expectedAPI))
			Expect(targetInfoArgs.ApiVersion).To(Equal(expectedAPIVersion))
			Expect(targetInfoArgs.Auth).To(Equal(expectedAuth))
			Expect(targetInfoArgs.MinCLIVersion).To(Equal(expectedMinCLIVersion))
			Expect(targetInfoArgs.Doppler).To(Equal(expectedDoppler))
			Expect(targetInfoArgs.Routing).To(Equal(expectedRouting))
			Expect(targetInfoArgs.SkipSSLValidation).To(Equal(skipSSLValidation))
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

		When("setting the same API and skip SSL configuration", func() {
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

})
