package v3action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient)
	})

	Describe("ResetSpaceIsolationSegment", func() {
		Context("when the organization does not have a default isolation segment", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.AssignSpaceToIsolationSegmentReturns(
					ccv3.Relationship{GUID: ""},
					ccv3.Warnings{"warning-1", "warning-2"}, nil)
			})

			It("returns an empty isolation segment GUID", func() {
				newIsolationSegmentName, warnings, err := actor.ResetSpaceIsolationSegment("some-org-guid", "some-space-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(newIsolationSegmentName).To(BeEmpty())

				Expect(fakeCloudControllerClient.AssignSpaceToIsolationSegmentCallCount()).To(Equal(1))
				spaceGUID, isolationSegmentGUID := fakeCloudControllerClient.AssignSpaceToIsolationSegmentArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(isolationSegmentGUID).To(BeEmpty())

				Expect(fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentCallCount()).To(Equal(1))
				orgGUID := fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentArgsForCall(0)
				Expect(orgGUID).To(Equal("some-org-guid"))

				Expect(fakeCloudControllerClient.GetIsolationSegmentCallCount()).To(Equal(0))
			})
		})

		Context("when the organization has a default isolation segment", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.AssignSpaceToIsolationSegmentReturns(
					ccv3.Relationship{GUID: ""},
					ccv3.Warnings{"warning-1", "warning-2"}, nil)
				fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentReturns(
					ccv3.Relationship{GUID: "some-iso-guid"},
					ccv3.Warnings{"warning-3", "warning-4"}, nil)
				fakeCloudControllerClient.GetIsolationSegmentReturns(
					ccv3.IsolationSegment{Name: "some-iso-name"},
					ccv3.Warnings{"warning-5", "warning-6"}, nil)
			})

			It("returns the org's isolation segment GUID", func() {
				newIsolationSegmentName, warnings, err := actor.ResetSpaceIsolationSegment("some-org-guid", "some-space-guid")

				Expect(fakeCloudControllerClient.AssignSpaceToIsolationSegmentCallCount()).To(Equal(1))
				spaceGUID, isolationSegmentGUID := fakeCloudControllerClient.AssignSpaceToIsolationSegmentArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(isolationSegmentGUID).To(BeEmpty())

				Expect(fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentCallCount()).To(Equal(1))
				orgGUID := fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentArgsForCall(0)
				Expect(orgGUID).To(Equal("some-org-guid"))

				Expect(fakeCloudControllerClient.GetIsolationSegmentCallCount()).To(Equal(1))
				isoSegGUID := fakeCloudControllerClient.GetIsolationSegmentArgsForCall(0)
				Expect(isoSegGUID).To(Equal("some-iso-guid"))

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6"))
				Expect(newIsolationSegmentName).To(Equal("some-iso-name"))
			})
		})

		Context("when assigning the space returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeCloudControllerClient.AssignSpaceToIsolationSegmentReturns(
					ccv3.Relationship{GUID: ""},
					ccv3.Warnings{"warning-1", "warning-2"}, expectedErr)
			})

			It("returns warnings and the error", func() {
				_, warnings, err := actor.ResetSpaceIsolationSegment("some-org-guid", "some-space-guid")
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(err).To(MatchError(expectedErr))
			})
		})

		Context("when getting the org's default isolation segments returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeCloudControllerClient.AssignSpaceToIsolationSegmentReturns(
					ccv3.Relationship{GUID: ""},
					ccv3.Warnings{"warning-1", "warning-2"}, nil)
				fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentReturns(
					ccv3.Relationship{GUID: "some-iso-guid"},
					ccv3.Warnings{"warning-3", "warning-4"}, expectedErr)
			})

			It("returns the warnings and an error", func() {
				_, warnings, err := actor.ResetSpaceIsolationSegment("some-org-guid", "some-space-guid")
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
				Expect(err).To(MatchError(expectedErr))
			})
		})

		Context("when getting the isolation segment returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeCloudControllerClient.AssignSpaceToIsolationSegmentReturns(
					ccv3.Relationship{GUID: ""},
					ccv3.Warnings{"warning-1", "warning-2"}, nil)
				fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentReturns(
					ccv3.Relationship{GUID: "some-iso-guid"},
					ccv3.Warnings{"warning-3", "warning-4"}, nil)
				fakeCloudControllerClient.GetIsolationSegmentReturns(
					ccv3.IsolationSegment{Name: "some-iso-name"},
					ccv3.Warnings{"warning-5", "warning-6"}, expectedErr)
			})

			It("returns the warnings and an error", func() {
				_, warnings, err := actor.ResetSpaceIsolationSegment("some-org-guid", "some-space-guid")
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6"))
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})
})
