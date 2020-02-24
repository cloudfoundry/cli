package v3action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LogCacheURL", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
		fakeConfig                *v3actionfakes.FakeConfig
		actualLogCacheURL         string
	)

	JustBeforeEach(func() {
		actualLogCacheURL = actor.LogCacheURL()
	})

	BeforeEach(func() {
		actor, fakeCloudControllerClient, fakeConfig, _, _ = NewTestActor()
		fakeConfig.TargetReturns("https://api.the-best-domain.com")
	})

	When("fakeCloudControllerClient.GetInfo() succeeds", func() {
		When("there is a log cache url", func() {
			var configuredLogcacheURL string
			BeforeEach(func() {
				configuredLogcacheURL = "https://log-cache.up-to-date-logging.com"
				fakeCloudControllerClient.GetInfoReturns(ccv3.Info{
					Links: ccv3.InfoLinks{
						LogCache: ccv3.APILink{
							HREF: configuredLogcacheURL,
						}}}, ccv3.ResourceLinks{}, ccv3.Warnings{}, nil)
			})
			It("uses it", func() {
				Expect(actualLogCacheURL).To(Equal(configuredLogcacheURL))
			})
		})

		When("there is no log cache url", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetInfoReturns(ccv3.Info{
					Links: ccv3.InfoLinks{},
				}, ccv3.ResourceLinks{}, ccv3.Warnings{}, nil)
			})
			It("uses the target", func() {
				if ccversion.MinSupportedV2ClientVersion != "2.128.0" {
					Fail("TIMEBOMB: This log-cache behavior should be removed in January 2021 when we no longer support cf-deployment 7.0.0")
				}
				Expect(actualLogCacheURL).To(Equal("https://log-cache.the-best-domain.com"))
			})
		})
	})

	When("fakeCloudControllerClient.GetInfo() fails", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetInfoReturns(ccv3.Info{
				Links: ccv3.InfoLinks{},
			}, ccv3.ResourceLinks{}, ccv3.Warnings{}, errors.New("awf splatz!"))
		})
		It("uses the target", func() {
			Expect(actualLogCacheURL).To(Equal("https://log-cache.the-best-domain.com"))
		})
	})
})
