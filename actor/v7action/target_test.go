package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Targeting", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		fakeConfig                *v7actionfakes.FakeConfig

		settings          TargetSettings
		skipSSLValidation bool
		targetedURL       string
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, fakeConfig, _, _, _, _ = NewTestActor()
	})

	Describe("SetTarget", func() {
		var (
			expectedAPI        string
			expectedAPIVersion string
			expectedAuth       string
			expectedDoppler    string
			expectedLogCache   string
			expectedRouting    string

			err      error
			warnings Warnings
		)

		BeforeEach(func() {
			expectedAPI = "https://api.foo.com"
			expectedAPIVersion = "3.81.0"
			expectedAuth = "https://login.foo.com"
			expectedDoppler = "wss://doppler.foo.com"
			expectedLogCache = "https://log-cache.foo.com"
			expectedRouting = "https://api.foo.com/routing"

			skipSSLValidation = true
			targetedURL = expectedAPI
			var meta struct {
				Version            string `json:"version"`
				HostKeyFingerprint string `json:"host_key_fingerprint"`
				OAuthClient        string `json:"oath_client"`
			}
			meta.Version = expectedAPIVersion

			rootResponse := ccv3.Info{
				Links: ccv3.InfoLinks{
					CCV3: resources.APILink{
						Meta: meta,
					},
					Logging: resources.APILink{
						HREF: expectedDoppler,
					},
					LogCache: resources.APILink{
						HREF: expectedLogCache,
					},
					Routing: resources.APILink{
						HREF: expectedRouting,
					},
					Login: resources.APILink{
						HREF: expectedAuth,
					},
				},
			}
			fakeCloudControllerClient.GetInfoReturns(rootResponse, ccv3.Warnings{"info-warning"}, nil)
		})

		JustBeforeEach(func() {
			settings = TargetSettings{
				SkipSSLValidation: skipSSLValidation,
				URL:               targetedURL,
			}
			warnings, err = actor.SetTarget(settings)
		})

		It("targets CF with the expected arguments", func() {
			Expect(fakeCloudControllerClient.TargetCFCallCount()).To(Equal(1))
			connectionSettings := fakeCloudControllerClient.TargetCFArgsForCall(0)
			Expect(connectionSettings.URL).To(Equal(expectedAPI))
			Expect(connectionSettings.SkipSSLValidation).To(BeTrue())
		})

		When("getting root info fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetInfoReturns(ccv3.Info{}, ccv3.Warnings{"info-warning"}, errors.New("info-error"))
			})

			It("returns an error and all warnings", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("info-error"))
				Expect(warnings).To(ConsistOf(Warnings{"info-warning"}))

				Expect(fakeCloudControllerClient.TargetCFCallCount()).To(Equal(1))
			})
		})

		It("sets all the target information", func() {
			Expect(fakeConfig.SetTargetInformationCallCount()).To(Equal(1))
			targetInfoArgs := fakeConfig.SetTargetInformationArgsForCall(0)

			Expect(targetInfoArgs.Api).To(Equal(expectedAPI))
			Expect(targetInfoArgs.ApiVersion).To(Equal(expectedAPIVersion))
			Expect(targetInfoArgs.Auth).To(Equal(expectedAuth))
			Expect(targetInfoArgs.Doppler).To(Equal(expectedDoppler))
			Expect(targetInfoArgs.LogCache).To(Equal(expectedLogCache))
			Expect(targetInfoArgs.Routing).To(Equal(expectedRouting))
			Expect(targetInfoArgs.SkipSSLValidation).To(Equal(skipSSLValidation))
			Expect(targetInfoArgs.CFOnK8s).To(BeFalse())
		})

		It("clears all the token information", func() {
			Expect(fakeConfig.SetTokenInformationCallCount()).To(Equal(1))
			accessToken, refreshToken, sshOAuthClient := fakeConfig.SetTokenInformationArgsForCall(0)

			Expect(accessToken).To(BeEmpty())
			Expect(refreshToken).To(BeEmpty())
			Expect(sshOAuthClient).To(BeEmpty())
		})

		It("clears the Kubernetes auth-info", func() {
			Expect(fakeConfig.SetKubernetesAuthInfoCallCount()).To(Equal(1))
			authInfo := fakeConfig.SetKubernetesAuthInfoArgsForCall(0)

			Expect(authInfo).To(BeEmpty())
		})

		It("succeeds and returns all warnings", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(ConsistOf(Warnings{"info-warning"}))
		})

		When("deployed on Kubernetes", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetInfoReturns(ccv3.Info{CFOnK8s: true}, nil, nil)
			})

			It("sets the CFOnK8s target information", func() {
				Expect(fakeConfig.SetTargetInformationCallCount()).To(Equal(1))
				targetInfoArgs := fakeConfig.SetTargetInformationArgsForCall(0)
				Expect(targetInfoArgs.CFOnK8s).To(BeTrue())
			})
		})
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
			Expect(targetInfoArgs.LogCache).To(BeEmpty())
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

		It("clears the Kubernetes auth-info", func() {
			actor.ClearTarget()

			Expect(fakeConfig.SetKubernetesAuthInfoCallCount()).To(Equal(1))
			authInfo := fakeConfig.SetKubernetesAuthInfoArgsForCall(0)

			Expect(authInfo).To(BeEmpty())
		})
	})
})
