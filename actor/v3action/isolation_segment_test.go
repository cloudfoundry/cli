package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Isolation Segment Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
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
				warnings, err := actor.CreateIsolationSegmentByName(IsolationSegment{Name: "some-isolation-segment"})
				Expect(err).ToNot(HaveOccurred())

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.CreateIsolationSegmentCallCount()).To(Equal(1))
				isolationSegmentName := fakeCloudControllerClient.CreateIsolationSegmentArgsForCall(0)
				Expect(isolationSegmentName).To(Equal(ccv3.IsolationSegment{Name: "some-isolation-segment"}))
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
					warnings, err := actor.CreateIsolationSegmentByName(IsolationSegment{Name: "some-isolation-segment"})
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			Context("when an UnprocessableEntityError occurs", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateIsolationSegmentReturns(
						ccv3.IsolationSegment{},
						ccv3.Warnings{"warning-1", "warning-2"},
						ccerror.UnprocessableEntityError{},
					)
				})

				It("returns an IsolationSegmentAlreadyExistsError and all warnings", func() {
					warnings, err := actor.CreateIsolationSegmentByName(IsolationSegment{Name: "some-isolation-segment"})
					Expect(err).To(MatchError(actionerror.IsolationSegmentAlreadyExistsError{Name: "some-isolation-segment"}))
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
					Expect(fakeCloudControllerClient.GetIsolationSegmentsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-iso-seg"}},
					))

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
					expectedErr = actionerror.OrganizationNotFoundError{Name: "some-org"}
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
				expectedErr = actionerror.IsolationSegmentNotFoundError{Name: "some-iso-seg"}
				fakeCloudControllerClient.GetIsolationSegmentsReturns(nil, ccv3.Warnings{"get-iso-warning"}, expectedErr)
			})

			It("returns the error", func() {
				warnings, err := actor.EntitleIsolationSegmentToOrganizationByName("some-iso-seg", "some-org")
				Expect(warnings).To(ConsistOf("get-iso-warning"))
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})

	Describe("AssignIsolationSegmentToSpaceByNameAndSpace", func() {
		Context("when the retrieving the isolation segment succeeds", func() {
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

			Context("when the assignment is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.AssignSpaceToIsolationSegmentReturns(ccv3.Relationship{GUID: "doesn't matter"}, ccv3.Warnings{"assignment-warnings-1", "assignment-warnings-2"}, nil)
				})

				It("returns the warnings", func() {
					warnings, err := actor.AssignIsolationSegmentToSpaceByNameAndSpace("some-iso-seg", "some-space-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("I r warnings", "I are two warnings", "assignment-warnings-1", "assignment-warnings-2"))

					Expect(fakeCloudControllerClient.GetIsolationSegmentsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetIsolationSegmentsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-iso-seg"}},
					))

					Expect(fakeCloudControllerClient.AssignSpaceToIsolationSegmentCallCount()).To(Equal(1))
					spaceGUID, isoGUID := fakeCloudControllerClient.AssignSpaceToIsolationSegmentArgsForCall(0)
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(isoGUID).To(Equal("some-iso-guid"))
				})
			})

			Context("when the assignment errors", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("foo bar")
					fakeCloudControllerClient.AssignSpaceToIsolationSegmentReturns(ccv3.Relationship{}, ccv3.Warnings{"assignment-warnings-1", "assignment-warnings-2"}, expectedErr)
				})

				It("returns the warnings and error", func() {
					warnings, err := actor.AssignIsolationSegmentToSpaceByNameAndSpace("some-iso-seg", "some-space-guid")
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("I r warnings", "I are two warnings", "assignment-warnings-1", "assignment-warnings-2"))
				})
			})
		})

		Context("when the retrieving the isolation segment errors", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("foo bar")
				fakeCloudControllerClient.GetIsolationSegmentsReturns(nil, ccv3.Warnings{"I r warnings", "I are two warnings"}, expectedErr)
			})

			It("returns the warnings and error", func() {
				warnings, err := actor.AssignIsolationSegmentToSpaceByNameAndSpace("some-iso-seg", "some-space-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("I r warnings", "I are two warnings"))
			})
		})
	})

	Describe("GetEffectiveIsolationSegmentBySpace", func() {
		Context("when the retrieving the space isolation segment succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceIsolationSegmentReturns(ccv3.Relationship{
					GUID: "some-iso-guid",
				}, ccv3.Warnings{"I r warnings", "I are two warnings"},
					nil,
				)
			})

			Context("when retrieving the isolation segment succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetIsolationSegmentReturns(ccv3.IsolationSegment{
						Name: "some-iso",
					},
						ccv3.Warnings{"iso-warnings-1", "iso-warnings-2"}, nil)
				})

				It("returns the warnings and IsolationSegment", func() {
					isolationSegment, warnings, err := actor.GetEffectiveIsolationSegmentBySpace("some-space-guid", "")
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("I r warnings", "I are two warnings", "iso-warnings-1", "iso-warnings-2"))
					Expect(isolationSegment).To(Equal(IsolationSegment{Name: "some-iso"}))

					Expect(fakeCloudControllerClient.GetSpaceIsolationSegmentCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSpaceIsolationSegmentArgsForCall(0)).To(Equal("some-space-guid"))

					Expect(fakeCloudControllerClient.GetIsolationSegmentCallCount()).To(Equal(1))
					arg := fakeCloudControllerClient.GetIsolationSegmentArgsForCall(0)
					Expect(arg).To(Equal("some-iso-guid"))
				})
			})

			Context("when retrieving the isolation segment errors", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("foo bar")
					fakeCloudControllerClient.GetIsolationSegmentReturns(ccv3.IsolationSegment{}, ccv3.Warnings{"iso-warnings-1", "iso-warnings-2"}, expectedErr)
				})

				It("returns the warnings and error", func() {
					_, warnings, err := actor.GetEffectiveIsolationSegmentBySpace("some-space-guid", "")
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("I r warnings", "I are two warnings", "iso-warnings-1", "iso-warnings-2"))
				})
			})

			Context("when the space does not have an isolation segment", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceIsolationSegmentReturns(ccv3.Relationship{
						GUID: "",
					}, ccv3.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				Context("when no org isolation segment is passed in", func() {
					It("returns NoRelationshipError", func() {
						_, warnings, err := actor.GetEffectiveIsolationSegmentBySpace("some-space-guid", "")
						Expect(err).To(MatchError(actionerror.NoRelationshipError{}))
						Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					})
				})

				Context("when an org default isolation segment is passed", func() {
					Context("when retrieving the isolation segment is successful", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetIsolationSegmentReturns(
								ccv3.IsolationSegment{
									Name: "some-iso-segment",
									GUID: "some-org-default-isolation-segment-guid",
								},
								ccv3.Warnings{"warning-3", "warning-4"},
								nil)
						})

						It("returns the org's default isolation segment", func() {
							isolationSegment, warnings, err := actor.GetEffectiveIsolationSegmentBySpace("some-space-guid", "some-org-default-isolation-segment-guid")
							Expect(isolationSegment).To(Equal(IsolationSegment{
								Name: "some-iso-segment",
								GUID: "some-org-default-isolation-segment-guid",
							}))
							Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
							Expect(err).ToNot(HaveOccurred())

							Expect(fakeCloudControllerClient.GetIsolationSegmentCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetIsolationSegmentArgsForCall(0)).To(Equal("some-org-default-isolation-segment-guid"))
						})
					})
				})
			})
		})

		Context("when the retrieving the space isolation segment errors", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("foo bar")
				fakeCloudControllerClient.GetSpaceIsolationSegmentReturns(ccv3.Relationship{}, ccv3.Warnings{"I r warnings", "I are two warnings"}, expectedErr)
			})

			It("returns the warnings and error", func() {
				_, warnings, err := actor.GetEffectiveIsolationSegmentBySpace("some-space-guid", "")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("I r warnings", "I are two warnings"))
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
				Expect(fakeCloudControllerClient.GetIsolationSegmentsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-iso-seg"}},
				))
			})
		})

		Context("when the isolation segment does *not* exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetIsolationSegmentsReturns(nil, ccv3.Warnings{"I r warnings", "I are two warnings"}, nil)
			})

			It("returns an IsolationSegmentNotFoundError", func() {
				_, warnings, err := actor.GetIsolationSegmentByName("some-iso-seg")
				Expect(err).To(MatchError(actionerror.IsolationSegmentNotFoundError{Name: "some-iso-seg"}))
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

	Describe("GetIsolationSegmentsByOrganization", func() {
		Context("when there are isolation segments entitled to this org", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetIsolationSegmentsReturns(
					[]ccv3.IsolationSegment{
						{Name: "some-iso-seg-1"},
						{Name: "some-iso-seg-2"},
					},
					ccv3.Warnings{"get isolation segments warning"},
					nil,
				)
			})

			It("returns the isolation segments and warnings", func() {
				isolationSegments, warnings, err := actor.GetIsolationSegmentsByOrganization("some-org-guid")
				Expect(err).ToNot(HaveOccurred())

				Expect(isolationSegments).To(ConsistOf(
					IsolationSegment{Name: "some-iso-seg-1"},
					IsolationSegment{Name: "some-iso-seg-2"},
				))
				Expect(warnings).To(ConsistOf("get isolation segments warning"))

				Expect(fakeCloudControllerClient.GetIsolationSegmentsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetIsolationSegmentsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{"some-org-guid"}},
				))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("some cc error")
				fakeCloudControllerClient.GetIsolationSegmentsReturns(
					[]ccv3.IsolationSegment{},
					ccv3.Warnings{"get isolation segments warning"},
					expectedError)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetIsolationSegmentsByOrganization("some-org-guid")
				Expect(warnings).To(ConsistOf("get isolation segments warning"))
				Expect(err).To(MatchError(expectedError))
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

	Describe("RevokeIsolationSegmentFromOrganizationByName", func() {
		Context("when the isolation segment exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetIsolationSegmentsReturns([]ccv3.IsolationSegment{
					{
						Name: "iso-1",
						GUID: "iso-1-guid-1",
					},
				}, ccv3.Warnings{"get-entitled-orgs-warning-1"}, nil)
			})

			Context("when the organization exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationsReturns([]ccv3.Organization{
						{
							Name: "org-1",
							GUID: "org-guid-1",
						},
					}, ccv3.Warnings{"get-orgs-warning-1"}, nil)
				})

				Context("when the revocation is successful", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.RevokeIsolationSegmentFromOrganizationReturns(ccv3.Warnings{"revoke-warnings-1"}, nil)
					})

					It("returns the warnings", func() {
						warnings, err := actor.RevokeIsolationSegmentFromOrganizationByName("iso-1", "org-1")
						Expect(err).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("get-entitled-orgs-warning-1", "get-orgs-warning-1", "revoke-warnings-1"))

						Expect(fakeCloudControllerClient.RevokeIsolationSegmentFromOrganizationCallCount()).To(Equal(1))
						isoGUID, orgGUID := fakeCloudControllerClient.RevokeIsolationSegmentFromOrganizationArgsForCall(0)
						Expect(isoGUID).To(Equal("iso-1-guid-1"))
						Expect(orgGUID).To(Equal("org-guid-1"))
					})
				})

				Context("when the revocation errors", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("Banana!")
						fakeCloudControllerClient.RevokeIsolationSegmentFromOrganizationReturns(ccv3.Warnings{"revoke-warnings-1"}, expectedErr)
					})

					It("from Organization", func() {
						warnings, err := actor.RevokeIsolationSegmentFromOrganizationByName("iso-1", "org-1")
						Expect(err).To(MatchError(expectedErr))
						Expect(warnings).To(ConsistOf("get-entitled-orgs-warning-1", "get-orgs-warning-1", "revoke-warnings-1"))
					})
				})
			})

			Context("when getting the organization errors", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationsReturns(nil, ccv3.Warnings{"get-orgs-warning-1"}, nil)
				})

				It("returns back the error", func() {
					warnings, err := actor.RevokeIsolationSegmentFromOrganizationByName("iso-1", "org-1")
					Expect(err).To(MatchError(actionerror.OrganizationNotFoundError{Name: "org-1"}))
					Expect(warnings).To(ConsistOf("get-entitled-orgs-warning-1", "get-orgs-warning-1"))

					Expect(fakeCloudControllerClient.RevokeIsolationSegmentFromOrganizationCallCount()).To(Equal(0))
				})
			})
		})

		Context("when getting the isolation segment errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetIsolationSegmentsReturns(nil, ccv3.Warnings{"get-entitled-orgs-warning-1"}, nil)
			})

			It("returns back the error", func() {
				warnings, err := actor.RevokeIsolationSegmentFromOrganizationByName("iso-2-org-1", "org-1")
				Expect(err).To(MatchError(actionerror.IsolationSegmentNotFoundError{Name: "iso-2-org-1"}))
				Expect(warnings).To(ConsistOf("get-entitled-orgs-warning-1"))

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(0))
			})
		})

	})

	Describe("SetOrganizationDefaultIsolationSegment", func() {
		Context("when the assignment is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PatchOrganizationDefaultIsolationSegmentReturns(
					ccv3.Relationship{GUID: "some-guid"},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns all warnings", func() {
				warnings, err := actor.SetOrganizationDefaultIsolationSegment("some-org-guid", "some-iso-seg-guid")
				Expect(err).ToNot(HaveOccurred())

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.PatchOrganizationDefaultIsolationSegmentCallCount()).To(Equal(1))
				orgGUID, isoSegGUID := fakeCloudControllerClient.PatchOrganizationDefaultIsolationSegmentArgsForCall(0)
				Expect(orgGUID).To(Equal("some-org-guid"))
				Expect(isoSegGUID).To(Equal("some-iso-seg-guid"))
			})
		})

		Context("when the assignment fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PatchOrganizationDefaultIsolationSegmentReturns(
					ccv3.Relationship{GUID: "some-guid"},
					ccv3.Warnings{"warning-1", "warning-2"},
					errors.New("some-error"),
				)
			})

			It("returns the error and all warnings", func() {
				warnings, err := actor.SetOrganizationDefaultIsolationSegment("some-org-guid", "some-iso-seg-guid")
				Expect(err).To(MatchError("some-error"))

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("ResetOrganizationDefaultIsolationSegment", func() {
		Context("when the assignment is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PatchOrganizationDefaultIsolationSegmentReturns(
					ccv3.Relationship{GUID: "some-guid"},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns all warnings", func() {
				warnings, err := actor.ResetOrganizationDefaultIsolationSegment("some-org-guid")
				Expect(err).ToNot(HaveOccurred())

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.PatchOrganizationDefaultIsolationSegmentCallCount()).To(Equal(1))
				orgGUID, isoSegGUID := fakeCloudControllerClient.PatchOrganizationDefaultIsolationSegmentArgsForCall(0)
				Expect(orgGUID).To(Equal("some-org-guid"))
				Expect(isoSegGUID).To(BeEmpty())
			})
		})

		Context("when the assignment fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PatchOrganizationDefaultIsolationSegmentReturns(
					ccv3.Relationship{GUID: "some-guid"},
					ccv3.Warnings{"warning-1", "warning-2"},
					errors.New("some-error"),
				)
			})

			It("returns the error and all warnings", func() {
				warnings, err := actor.ResetOrganizationDefaultIsolationSegment("some-org-guid")
				Expect(err).To(MatchError("some-error"))

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})
