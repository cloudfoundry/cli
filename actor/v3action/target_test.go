package v3action_test

import (
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Targeting", func() {
	var (
		actor             *Actor
		skipSSLValidation bool

		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
		fakeConfig                *v3actionfakes.FakeConfig
		settings                  TargetSettings
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v3actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil)

		settings = TargetSettings{
			SkipSSLValidation: skipSSLValidation,
		}
	})

	Describe("SetTarget", func() {
		var expectedAPI, expectedAPIVersion, expectedAuth, expectedDoppler, expectedRouting string

		BeforeEach(func() {
			expectedAPI = "https://api.foo.com"
			expectedAPIVersion = "2.59.0"
			expectedAuth = "https://login.foo.com"
			expectedDoppler = "wss://doppler.foo.com"
			expectedRouting = "https://api.foo.com/routing"

			settings.URL = expectedAPI
			var meta struct {
				Version            string `json:"version"`
				HostKeyFingerprint string `json:"host_key_fingerprint"`
				OAuthClient        string `json:"oath_client"`
			}
			meta.Version = expectedAPIVersion
			fakeCloudControllerClient.TargetCFReturns(ccv3.Info{
				Links: ccv3.InfoLinks{
					CCV3: ccv3.APILink{
						Meta: meta},
					Logging: ccv3.APILink{
						HREF: expectedDoppler,
					},
					Routing: ccv3.APILink{
						HREF: expectedRouting,
					},
					UAA: ccv3.APILink{
						HREF: expectedAuth,
					}}}, ccv3.Warnings{}, nil)
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
			api, apiVersion, auth, _, doppler, routing, sslDisabled := fakeConfig.SetTargetInformationArgsForCall(0)

			Expect(api).To(Equal(expectedAPI))
			Expect(apiVersion).To(Equal(expectedAPIVersion))
			Expect(auth).To(Equal(expectedAuth))
			Expect(doppler).To(Equal(expectedDoppler))
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

	Describe("ClearTarget", func() {
		It("clears all the target information", func() {
			actor.ClearTarget()
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
			actor.ClearTarget()

			Expect(fakeConfig.SetTokenInformationCallCount()).To(Equal(1))
			accessToken, refreshToken, sshOAuthClient := fakeConfig.SetTokenInformationArgsForCall(0)

			Expect(accessToken).To(BeEmpty())
			Expect(refreshToken).To(BeEmpty())
			Expect(sshOAuthClient).To(BeEmpty())
		})
	})
})
