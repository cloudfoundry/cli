package v3action_test

import (
	"errors"
	"net/url"

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
		Context("when the create is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateIsolationSegmentReturns(
					ccv3.IsolationSegment{},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns all warnings", func() {
				warnings, err := actor.CreateIsolationSegmentByName("some-isolation-segment-guid")
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
					warnings, err := actor.CreateIsolationSegmentByName("isolation-segment")
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

				It("returns an IsolationSegmentAlreadyExistsError and all warnings", func() {
					warnings, err := actor.CreateIsolationSegmentByName("isolation-segment")
					Expect(err).To(MatchError(IsolationSegmentAlreadyExistsError{Name: "isolation-segment"}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})
	})

	Describe("DeleteIsolationSegmentByName", func() {
		Context("when the isolation segment is found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetIsolationSegmentsReturns([]ccv3.IsolationSegment{
					{
						GUID: "some-iso-guid",
						Name: "some-iso-seg",
					},
				}, ccv3.Warnings{"I r warnings", "I are two warnings"},
					nil,
				)
			})

			Context("when the delete is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteIsolationSegmentReturns(ccv3.Warnings{"delete warning-1", "delete warning-2"}, nil)
				})

				It("returns back all warnings", func() {
					warnings, err := actor.DeleteIsolationSegmentByName("some-iso-seg")
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("I r warnings", "I are two warnings", "delete warning-1", "delete warning-2"))

					Expect(fakeCloudControllerClient.GetIsolationSegmentsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetIsolationSegmentsArgsForCall(0)).To(Equal(url.Values{ccv3.NameFilter: []string{"some-iso-seg"}}))

					Expect(fakeCloudControllerClient.DeleteIsolationSegmentCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.DeleteIsolationSegmentArgsForCall(0)).To(Equal("some-iso-guid"))
				})
			})

			Context("when the delete returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some-cc-error")
					fakeCloudControllerClient.DeleteIsolationSegmentReturns(ccv3.Warnings{"delete warning-1", "delete warning-2"}, expectedErr)
				})

				It("returns back the error and all warnings", func() {
					warnings, err := actor.DeleteIsolationSegmentByName("some-iso-seg")
					Expect(warnings).To(ConsistOf("I r warnings", "I are two warnings", "delete warning-1", "delete warning-2"))
					Expect(err).To(MatchError(expectedErr))
				})
			})
		})

		Context("when the search errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-cc-error")
				fakeCloudControllerClient.GetIsolationSegmentsReturns(nil, ccv3.Warnings{"I r warnings", "I are two warnings"}, expectedErr)
			})

			It("returns the error and all warnings", func() {
				warnings, err := actor.DeleteIsolationSegmentByName("some-iso-seg")
				Expect(warnings).To(ConsistOf("I r warnings", "I are two warnings"))
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})

	Describe("EntitleIsolationSegmentToOrganizationByName", func() {
		Context("when the isolation segment exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetIsolationSegmentsReturns([]ccv3.IsolationSegment{
					{
						Name: "some-iso-seg",
						GUID: "some-iso-guid",
					},
				}, ccv3.Warnings{"get-iso-warning"}, nil)
			})

			Context("when the organization exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationsReturns([]ccv3.Organization{
						{
							Name: "some-org",
							GUID: "some-org-guid",
						},
					}, ccv3.Warnings{"get-org-warning"}, nil)
				})

				Context("when the relationship succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.EntitleIsolationSegmentToOrganizationsReturns(
							ccv3.RelationshipList{GUIDs: []string{"some-relationship-guid"}},
							ccv3.Warnings{"entitle-iso-to-org-warning"},
							nil)
					})

					It("returns all warnings", func() {
						warnings, err := actor.EntitleIsolationSegmentToOrganizationByName("some-iso-seg", "some-org")
						Expect(warnings).To(ConsistOf("get-iso-warning", "get-org-warning", "entitle-iso-to-org-warning"))
						Expect(err).ToNot(HaveOccurred())
						Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetIsolationSegmentsCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.EntitleIsolationSegmentToOrganizationsCallCount()).To(Equal(1))
					})
				})

				Context("when the relationship fails", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("toxic-relationship")
						fakeCloudControllerClient.EntitleIsolationSegmentToOrganizationsReturns(
							ccv3.RelationshipList{},
							ccv3.Warnings{"entitle-iso-to-org-warning"},
							expectedErr)
					})

					It("returns the error", func() {
						warnings, err := actor.EntitleIsolationSegmentToOrganizationByName("some-iso-seg", "some-org")
						Expect(warnings).To(ConsistOf("get-iso-warning", "get-org-warning", "entitle-iso-to-org-warning"))
						Expect(err).To(MatchError(expectedErr))
					})

				})
			})

			Context("when retrieving the orgs errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = OrganizationNotFoundError{Name: "some-org"}
					fakeCloudControllerClient.GetOrganizationsReturns(nil, ccv3.Warnings{"get-org-warning"}, expectedErr)
				})

				It("returns the error", func() {
					warnings, err := actor.EntitleIsolationSegmentToOrganizationByName("some-iso-seg", "some-org")
					Expect(warnings).To(ConsistOf("get-org-warning", "get-iso-warning"))
					Expect(err).To(MatchError(expectedErr))
				})
			})
		})

		Context("when retrieving the isolation segment errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = IsolationSegmentNotFoundError{Name: "some-iso-seg"}
				fakeCloudControllerClient.GetIsolationSegmentsReturns(nil, ccv3.Warnings{"get-iso-warning"}, expectedErr)
			})

			It("returns the error", func() {
				warnings, err := actor.EntitleIsolationSegmentToOrganizationByName("some-iso-seg", "some-org")
				Expect(warnings).To(ConsistOf("get-iso-warning"))
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})

	Describe("GetIsolationSegmentByName", func() {
		Context("when the isolation segment exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetIsolationSegmentsReturns([]ccv3.IsolationSegment{
					{
						GUID: "some-iso-guid",
						Name: "some-iso-seg",
					},
				}, ccv3.Warnings{"I r warnings", "I are two warnings"},
					nil,
				)
			})

			It("returns the isolation segment and warnings", func() {
				segment, warnings, err := actor.GetIsolationSegmentByName("some-iso-seg")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("I r warnings", "I are two warnings"))
				Expect(segment).To(Equal(IsolationSegment{
					GUID: "some-iso-guid",
					Name: "some-iso-seg",
				}))

				Expect(fakeCloudControllerClient.GetIsolationSegmentsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetIsolationSegmentsArgsForCall(0)).To(Equal(url.Values{ccv3.NameFilter: []string{"some-iso-seg"}}))
			})
		})

		Context("when the isolation segment does *not* exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetIsolationSegmentsReturns(nil, ccv3.Warnings{"I r warnings", "I are two warnings"}, nil)
			})

			It("returns an IsolationSegmentNotFoundError", func() {
				_, warnings, err := actor.GetIsolationSegmentByName("some-iso-seg")
				Expect(err).To(MatchError(IsolationSegmentNotFoundError{Name: "some-iso-seg"}))
				Expect(warnings).To(ConsistOf("I r warnings", "I are two warnings"))
			})
		})

		Context("when the cloud controller errors", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some-cc-error")
				fakeCloudControllerClient.GetIsolationSegmentsReturns(nil, ccv3.Warnings{"I r warnings", "I are two warnings"}, expectedErr)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := actor.GetIsolationSegmentByName("some-iso-seg")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("I r warnings", "I are two warnings"))
			})
		})
	})

	Describe("GetIsolationSegmentSummaries", func() {
		Context("when getting isolation segments succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetIsolationSegmentsReturns([]ccv3.IsolationSegment{
					{
						Name: "iso-seg-1",
						GUID: "iso-guid-1",
					},
					{
						Name: "iso-seg-2",
						GUID: "iso-guid-2",
					},
				}, ccv3.Warnings{"get-iso-warning"}, nil)
			})

			Context("when getting entitled organizations succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetIsolationSegmentOrganizationsByIsolationSegmentReturnsOnCall(0, []ccv3.Organization{}, ccv3.Warnings{"get-entitled-orgs-warning-1"}, nil)
					fakeCloudControllerClient.GetIsolationSegmentOrganizationsByIsolationSegmentReturnsOnCall(1, []ccv3.Organization{
						{
							Name: "iso-2-org-1",
							GUID: "iso-2-org-guid-1",
						},
						{
							Name: "iso-2-org-2",
							GUID: "iso-2-org-guid-2",
						},
					}, ccv3.Warnings{"get-entitled-orgs-warning-2"}, nil)
				})

				It("returns all isolation segment summaries and all warnings", func() {
					isoSummaries, warnings, err := actor.GetIsolationSegmentSummaries()
					Expect(warnings).To(ConsistOf("get-iso-warning", "get-entitled-orgs-warning-1", "get-entitled-orgs-warning-2"))
					Expect(err).ToNot(HaveOccurred())
					Expect(isoSummaries).To(ConsistOf([]IsolationSegmentSummary{
						{
							Name:         "iso-seg-1",
							EntitledOrgs: []string{},
						},
						{
							Name:         "iso-seg-2",
							EntitledOrgs: []string{"iso-2-org-1", "iso-2-org-2"},
						},
					}))

					Expect(fakeCloudControllerClient.GetIsolationSegmentsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetIsolationSegmentsArgsForCall(0)).To(BeEmpty())
					Expect(fakeCloudControllerClient.GetIsolationSegmentOrganizationsByIsolationSegmentCallCount()).To(Equal(2))
					Expect(fakeCloudControllerClient.GetIsolationSegmentOrganizationsByIsolationSegmentArgsForCall(0)).To(Equal("iso-guid-1"))
					Expect(fakeCloudControllerClient.GetIsolationSegmentOrganizationsByIsolationSegmentArgsForCall(1)).To(Equal("iso-guid-2"))
				})
			})

			Context("when getting entitled organizations fails", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeCloudControllerClient.GetIsolationSegmentOrganizationsByIsolationSegmentReturns(nil, ccv3.Warnings{"get-entitled-orgs-warning"}, expectedErr)
				})

				It("returns the error and warnings", func() {
					_, warnings, err := actor.GetIsolationSegmentSummaries()
					Expect(warnings).To(ConsistOf("get-iso-warning", "get-entitled-orgs-warning"))
					Expect(err).To(MatchError(expectedErr))
				})
			})
		})

		Context("when getting isolation segments fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeCloudControllerClient.GetIsolationSegmentsReturns(nil, ccv3.Warnings{"get-iso-warning"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetIsolationSegmentSummaries()
				Expect(warnings).To(ConsistOf("get-iso-warning"))
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})
})
