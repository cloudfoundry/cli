package v3action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Isolation Segment Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient)
	})

	Describe("CreateIsolationSegment", func() {
		Context("when there are associated tasks", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateIsolationSegmentReturns(
					ccv3.IsolationSegment{},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns all tasks associated with the application and all warnings", func() {
				warnings, err := actor.CreateIsolationSegment("some-isolation-segment-guid")
				Expect(err).ToNot(HaveOccurred())

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.CreateIsolationSegmentCallCount()).To(Equal(1))
				isolationSegmentName := fakeCloudControllerClient.CreateIsolationSegmentArgsForCall(0)
				Expect(isolationSegmentName).To(Equal("some-isolation-segment-guid"))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			Context("when an unexpected error occurs", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I am a CloudControllerClient Error")
					fakeCloudControllerClient.CreateIsolationSegmentReturns(
						ccv3.IsolationSegment{},
						ccv3.Warnings{"warning-1", "warning-2"},
						expectedErr,
					)
				})

				It("returns the same error and all warnings", func() {
					warnings, err := actor.CreateIsolationSegment("isolation-segment")
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			Context("when an UnprocessableEntityError occurs", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateIsolationSegmentReturns(
						ccv3.IsolationSegment{},
						ccv3.Warnings{"warning-1", "warning-2"},
						cloudcontroller.UnprocessableEntityError{},
					)
				})

				It("returns the same error and all warnings", func() {
					warnings, err := actor.CreateIsolationSegment("isolation-segment")
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})
	})
})
