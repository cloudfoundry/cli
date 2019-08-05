package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeConfig := new(v7actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil, nil)
	})

	Describe("CreateSpace", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			_, warnings, executeErr = actor.CreateSpace("space-name", "org-guid")
		})

		When("the API layer calls are successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateSpaceReturns(
					ccv3.Space{GUID: "space-guid", Name: "space-name"},
					ccv3.Warnings{"create-warning-1", "create-warning-2"},
					nil)
			})
		})

		When("the cc client returns an NameNotUniqueInOrgError", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateSpaceReturns(
					ccv3.Space{},
					ccv3.Warnings{"create-space-warning"},
					ccerror.NameNotUniqueInOrgError{},
				)
			})

			It("returns the SpaceAlreadyExistsError and warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.SpaceAlreadyExistsError{
					Space: "space-name",
				}))
				Expect(warnings).To(ConsistOf("create-space-warning"))
			})
		})

		When("the cc client returns a different error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateSpaceReturns(
					ccv3.Space{},
					ccv3.Warnings{"warning"},
					errors.New("api-error"),
				)
			})

			It("it returns an error and prints warnings", func() {
				Expect(warnings).To(ConsistOf("warning"))
				Expect(executeErr).To(MatchError("api-error"))

				Expect(fakeCloudControllerClient.CreateSpaceCallCount()).To(Equal(1))
			})
		})
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

	Describe("GetOrganizationSpaces", func() {
		When("there are spaces in the org", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv3.Space{
						{
							GUID: "space-1-guid",
							Name: "space-1",
						},
						{
							GUID: "space-2-guid",
							Name: "space-2",
						},
					},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("returns all spaces and all warnings", func() {
				spaces, warnings, err := actor.GetOrganizationSpaces("some-org-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(spaces).To(Equal(
					[]Space{
						{
							GUID: "space-1-guid",
							Name: "space-1",
						},
						{
							GUID: "space-2-guid",
							Name: "space-2",
						},
					}))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal(
					[]ccv3.Query{
						{
							Key:    ccv3.OrganizationGUIDFilter,
							Values: []string{"some-org-guid"},
						},
					}))
			})
		})

		When("an error is encountered", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("cc-get-spaces-error")
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv3.Space{},
					ccv3.Warnings{"warning-1", "warning-2"},
					returnedErr,
				)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := actor.GetOrganizationSpaces("some-org-guid")

				Expect(err).To(MatchError(returnedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("DeleteSpaceByNameAndOrganizationName", func() {
		var (
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			warnings, err = actor.DeleteSpaceByNameAndOrganizationName("some-space", "some-org")
		})

		When("the org is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv3.Organization{},
					ccv3.Warnings{
						"warning-1",
						"warning-2",
					},
					nil,
				)
			})

			It("returns an OrganizationNotFoundError", func() {
				Expect(err).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("the org is found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv3.Organization{{Name: "some-org", GUID: "some-org-guid"}},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			When("the space is not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv3.Space{},
						ccv3.Warnings{"warning-3", "warning-4"},
						nil,
					)
				})

				It("returns an SpaceNotFoundError", func() {
					Expect(err).To(MatchError(actionerror.SpaceNotFoundError{Name: "some-space"}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
				})
			})

			When("the space is found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv3.Space{{GUID: "some-space-guid"}},
						ccv3.Warnings{"warning-3", "warning-4"},
						nil,
					)
				})

				When("the delete returns an error", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("some delete space error")
						fakeCloudControllerClient.DeleteSpaceReturns(
							ccv3.JobURL(""),
							ccv3.Warnings{"warning-5", "warning-6"},
							expectedErr,
						)
					})

					It("returns the error", func() {
						Expect(err).To(Equal(expectedErr))
						Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6"))
					})
				})

				When("the delete returns a job", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.DeleteSpaceReturns(
							ccv3.JobURL("some-url"),
							ccv3.Warnings{"warning-5", "warning-6"},
							nil,
						)
					})

					When("polling errors", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("Never expected, by anyone")
							fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"warning-7", "warning-8"}, expectedErr)
						})

						It("returns the error", func() {
							Expect(err).To(Equal(expectedErr))
							Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6", "warning-7", "warning-8"))
						})
					})

					When("the job is successful", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"warning-7", "warning-8"}, nil)
						})

						It("returns warnings and no error", func() {
							Expect(err).ToNot(HaveOccurred())
							Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6", "warning-7", "warning-8"))

							Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(Equal([]ccv3.Query{{
								Key:    ccv3.NameFilter,
								Values: []string{"some-org"},
							}}))

							Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal([]ccv3.Query{{
								Key:    ccv3.NameFilter,
								Values: []string{"some-space"},
							},
								{
									Key:    ccv3.OrganizationGUIDFilter,
									Values: []string{"some-org-guid"},
								},
							}))

							Expect(fakeCloudControllerClient.DeleteSpaceCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.DeleteSpaceArgsForCall(0)).To(Equal("some-space-guid"))

							Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.PollJobArgsForCall(0)).To(Equal(ccv3.JobURL("some-url")))
						})
					})
				})
			})
		})
	})
})
