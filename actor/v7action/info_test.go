package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Info Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("GetLogCacheEndpoint", func() {
		When("getting info is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetInfoReturns(
					ccv3.Info{
						Links: ccv3.InfoLinks{
							LogCache: ccv3.APILink{HREF: "some-log-cache-url"},
						},
					},
					nil,
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns all warnings and the log cache url", func() {
				logCacheURL, warnings, err := actor.GetLogCacheEndpoint()
				Expect(err).ToNot(HaveOccurred())

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetInfoCallCount()).To(Equal(1))
				Expect(logCacheURL).To(Equal("some-log-cache-url"))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetInfoReturns(
					ccv3.Info{},
					nil,
					ccv3.Warnings{"warning-1", "warning-2"},
					expectedErr,
				)
			})

			It("returns the same error and all warnings", func() {
				_, warnings, err := actor.GetLogCacheEndpoint()
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})

		})
	})
})
