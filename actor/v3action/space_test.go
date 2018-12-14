package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		fakeConfig := new(v3actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil)
	})

	Describe("ResetSpaceIsolationSegment", func() {
		When("the organization does not have a default isolation segment", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateSpaceIsolationSegmentRelationshipReturns(
					ccv3.Relationship{GUID: ""},
					ccv3.Warnings{"warning-1", "warning-2"}, nil)
			})

			It("returns an empty isolation segment GUID", func() {
				newIsolationSegmentName, warnings, err := actor.ResetSpaceIsolationSegment("some-org-guid", "some-space-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(newIsolationSegmentName).To(BeEmpty())

				Expect(fakeCloudControllerClient.UpdateSpaceIsolationSegmentRelationshipCallCount()).To(Equal(1))
				spaceGUID, isolationSegmentGUID := fakeCloudControllerClient.UpdateSpaceIsolationSegmentRelationshipArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(isolationSegmentGUID).To(BeEmpty())

				Expect(fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentCallCount()).To(Equal(1))
				orgGUID := fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentArgsForCall(0)
				Expect(orgGUID).To(Equal("some-org-guid"))

				Expect(fakeCloudControllerClient.GetIsolationSegmentCallCount()).To(Equal(0))
			})
		})

		When("the organization has a default isolation segment", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateSpaceIsolationSegmentRelationshipReturns(
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

				Expect(fakeCloudControllerClient.UpdateSpaceIsolationSegmentRelationshipCallCount()).To(Equal(1))
				spaceGUID, isolationSegmentGUID := fakeCloudControllerClient.UpdateSpaceIsolationSegmentRelationshipArgsForCall(0)
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

		When("assigning the space returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeCloudControllerClient.UpdateSpaceIsolationSegmentRelationshipReturns(
					ccv3.Relationship{GUID: ""},
					ccv3.Warnings{"warning-1", "warning-2"}, expectedErr)
			})

			It("returns warnings and the error", func() {
				_, warnings, err := actor.ResetSpaceIsolationSegment("some-org-guid", "some-space-guid")
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(err).To(MatchError(expectedErr))
			})
		})

		When("getting the org's default isolation segments returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeCloudControllerClient.UpdateSpaceIsolationSegmentRelationshipReturns(
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

		When("getting the isolation segment returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeCloudControllerClient.UpdateSpaceIsolationSegmentRelationshipReturns(
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

	Describe("GetSpaceByNameAndOrganization", func() {
		var (
			spaceName string
			orgGUID   string

			space      Space
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			spaceName = "some-space"
			orgGUID = "some-org-guid"
		})

		JustBeforeEach(func() {
			space, warnings, executeErr = actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
		})

		When("the GetSpace call is successful", func() {
			When("the cloud controller returns back one space", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv3.Space{{GUID: "some-space-guid", Name: spaceName}},
						ccv3.Warnings{"some-space-warning"}, nil)
				})

				It("returns back the first space and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(space).To(Equal(Space{
						GUID: "some-space-guid",
						Name: spaceName,
					}))
					Expect(warnings).To(ConsistOf("some-space-warning"))

					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{spaceName}},
						ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
					))
				})
			})

			When("the cloud controller returns back no spaces", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						nil, ccv3.Warnings{"some-space-warning"}, nil)
				})

				It("returns a SpaceNotFoundError and warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: spaceName}))

					Expect(warnings).To(ConsistOf("some-space-warning"))
				})
			})
		})

		When("the GetSpace call is unsuccessful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					nil,
					ccv3.Warnings{"some-space-warning"},
					errors.New("cannot get space"))
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError("cannot get space"))
				Expect(warnings).To(ConsistOf("some-space-warning"))
			})
		})

	})

	Describe("GetSpacesByGUIDs", func() {
		When("the GetSpace call is successful", func() {
			When("the cloud controller returns back multiple spaces", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv3.Space{
							{
								GUID: "space-guid-1",
								Name: "space-1",
								Relationships: ccv3.Relationships{
									constant.RelationshipTypeOrganization: ccv3.Relationship{GUID: "org-guid-1"},
								},
							},
							{
								GUID: "space-guid-2",
								Name: "space-2",
								Relationships: ccv3.Relationships{
									constant.RelationshipTypeOrganization: ccv3.Relationship{GUID: "org-guid-2"},
								},
							},
						},
						ccv3.Warnings{"space-warning-1", "space-warning-2"}, nil)
				})

				It("returns back the first space and warnings", func() {
					spaces, warnings, executeErr := actor.GetSpacesByGUIDs("space-guid-1", "space-guid-2")
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(spaces).To(ConsistOf(
						Space{
							GUID:             "space-guid-1",
							Name:             "space-1",
							OrganizationGUID: "org-guid-1",
						},
						Space{
							GUID:             "space-guid-2",
							Name:             "space-2",
							OrganizationGUID: "org-guid-2",
						},
					))
					Expect(warnings).To(ConsistOf("space-warning-1", "space-warning-2"))

					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{"space-guid-1", "space-guid-2"}},
					))
				})
			})
		})

		When("the GetSpace call is unsuccessful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					nil,
					ccv3.Warnings{"space-warning-1", "space-warning-2"},
					errors.New("cannot get space"))
			})

			It("returns an error and warnings", func() {
				_, warnings, executeErr := actor.GetSpacesByGUIDs("space-guid-1", "space-guid-2")
				Expect(executeErr).To(MatchError("cannot get space"))
				Expect(warnings).To(ConsistOf("space-warning-1", "space-warning-2"))
			})
		})
	})
})
