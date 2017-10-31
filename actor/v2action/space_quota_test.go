package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpaceQuota Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetSpaceQuota", func() {
		Context("when the space quota exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotaReturns(
					ccv2.SpaceQuota{
						GUID: "some-space-quota-guid",
						Name: "some-space-quota",
					},
					ccv2.Warnings{"warning-1"},
					nil,
				)
			})

			It("returns the space quota and warnings", func() {
				spaceQuota, warnings, err := actor.GetSpaceQuota("some-space-quota-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(spaceQuota).To(Equal(SpaceQuota{
					GUID: "some-space-quota-guid",
					Name: "some-space-quota",
				}))
				Expect(warnings).To(ConsistOf("warning-1"))

				Expect(fakeCloudControllerClient.GetSpaceQuotaCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpaceQuotaArgsForCall(0)).To(Equal(
					"some-space-quota-guid"))
			})
		})

		Context("when the space quota does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotaReturns(ccv2.SpaceQuota{}, nil, ccerror.ResourceNotFoundError{})
			})

			It("returns an SpaceQuotaNotFoundError", func() {
				_, _, err := actor.GetSpaceQuota("some-space-quota-guid")
				Expect(err).To(MatchError(actionerror.SpaceQuotaNotFoundError{GUID: "some-space-quota-guid"}))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some space quota error")
				fakeCloudControllerClient.GetSpaceQuotaReturns(ccv2.SpaceQuota{}, ccv2.Warnings{"warning-1", "warning-2"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetSpaceQuota("some-space-quota-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})
