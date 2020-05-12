package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
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
			space      Space
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			space, warnings, executeErr = actor.CreateSpace("space-name", "org-guid")
		})

		When("the API layer calls are successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateSpaceReturns(
					ccv3.Space{GUID: "space-guid", Name: "space-name"},
					ccv3.Warnings{"not-fatal-warning"},
					nil)
			})

			It("creates a space successfully", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(space.Name).To(Equal("space-name"))
				Expect(space.GUID).To(Equal("space-guid"))
				Expect(warnings).To(ConsistOf("not-fatal-warning"))
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
					resources.Relationship{GUID: ""},
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
					resources.Relationship{GUID: ""},
					ccv3.Warnings{"warning-1", "warning-2"}, nil)
				fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentReturns(
					resources.Relationship{GUID: "some-iso-guid"},
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
					resources.Relationship{GUID: ""},
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
					resources.Relationship{GUID: ""},
					ccv3.Warnings{"warning-1", "warning-2"}, nil)
				fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentReturns(
					resources.Relationship{GUID: "some-iso-guid"},
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
					resources.Relationship{GUID: ""},
					ccv3.Warnings{"warning-1", "warning-2"}, nil)
				fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentReturns(
					resources.Relationship{GUID: "some-iso-guid"},
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
						[]ccv3.Space{
							{
								GUID: "some-space-guid",
								Name: spaceName,
							},
						},
						ccv3.IncludedResources{},
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

			When("the cloud controller returns a space with a quota relationship", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv3.Space{{GUID: "some-space-guid", Name: spaceName,
							Relationships: resources.Relationships{
								constant.RelationshipTypeQuota: resources.Relationship{GUID: "some-space-quota-guid"},
							}}},
						ccv3.IncludedResources{},
						ccv3.Warnings{"some-space-warning"}, nil)
				})
				It("returns the quota relationship", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))

					Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{spaceName}},
						ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
					))
					Expect(space).To(Equal(Space{
						GUID: "some-space-guid",
						Name: spaceName,
						Relationships: resources.Relationships{
							constant.RelationshipTypeQuota: resources.Relationship{GUID: "some-space-quota-guid"},
						},
					}))
				})

			})

			When("the cloud controller returns back no spaces", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						nil, ccv3.IncludedResources{}, ccv3.Warnings{"some-space-warning"}, nil)
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
					ccv3.IncludedResources{},
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
					ccv3.IncludedResources{},
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
						{Key: ccv3.OrganizationGUIDFilter, Values: []string{"some-org-guid"}},
						{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
					}))
			})

			When("a label selector is provided", func() {

				It("passes the label selector through", func() {
					_, _, err := actor.GetOrganizationSpacesWithLabelSelector("some-org-guid", "some-label-selector")
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))

					expectedQuery := []ccv3.Query{
						{Key: ccv3.OrganizationGUIDFilter, Values: []string{"some-org-guid"}},
						{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
						{Key: ccv3.LabelSelectorFilter, Values: []string{"some-label-selector"}},
					}
					actualQuery := fakeCloudControllerClient.GetSpacesArgsForCall(0)
					Expect(actualQuery).To(Equal(expectedQuery))
				})

			})
		})

		When("an error is encountered", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("cc-get-spaces-error")
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv3.Space{},
					ccv3.IncludedResources{},
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
						ccv3.IncludedResources{},
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
						ccv3.IncludedResources{},
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

	Describe("RenameSpaceByNameAndOrganizationGUID", func() {
		var (
			oldSpaceName string
			newSpaceName string
			orgGUID      string

			space      Space
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			space, warnings, executeErr = actor.RenameSpaceByNameAndOrganizationGUID(
				oldSpaceName,
				newSpaceName,
				orgGUID,
			)
		})

		It("delegate to the actor to get the space", func() {
			// assert on the underlying client call because we dont have a fake actor
			Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
		})

		When("getting the space fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					nil,
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-space-warning"},
					errors.New("get-space-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("get-space-error"))
				Expect(warnings).To(ConsistOf("get-space-warning"))
			})
		})

		When("getting the space succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv3.Space{{Name: oldSpaceName, GUID: "space-guid"}},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-space-warning"},
					nil,
				)
			})

			It("delegates to the client to update the space", func() {
				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateSpaceCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateSpaceArgsForCall(0)).To(Equal(ccv3.Space{
					GUID: "space-guid",
					Name: newSpaceName,
				}))
			})

			When("updating the space fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateSpaceReturns(
						ccv3.Space{},
						ccv3.Warnings{"update-space-warning"},
						errors.New("update-space-error"),
					)
				})

				It("returns an error and all warnings", func() {
					Expect(executeErr).To(MatchError("update-space-error"))
					Expect(warnings).To(ConsistOf("get-space-warning", "update-space-warning"))
				})

			})

			When("updating the space succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateSpaceReturns(
						ccv3.Space{Name: newSpaceName, GUID: "space-guid"},
						ccv3.Warnings{"update-space-warning"},
						nil,
					)
				})

				It("returns warnings and no error", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-space-warning", "update-space-warning"))
					Expect(space).To(Equal(Space{Name: newSpaceName, GUID: "space-guid"}))
				})
			})
		})
	})

	Describe("GetSpaceSummaryByNameAndOrganization", func() {
		var (
			spaceSummary SpaceSummary
			warnings     Warnings
			err          error

			org        ccv3.Organization
			ccv3Spaces []ccv3.Space
			apps       []resources.Application
		)

		JustBeforeEach(func() {
			spaceSummary, warnings, err = actor.GetSpaceSummaryByNameAndOrganization("space-name", "org-guid")
		})

		BeforeEach(func() {
			org = ccv3.Organization{GUID: "some-org-guid", Name: "some-org-name"}

			ccv3Spaces = []ccv3.Space{
				{
					GUID: "some-space-guid",
					Name: "some-space-name",
				},
			}
			fakeCloudControllerClient.GetSpacesReturns(ccv3Spaces, ccv3.IncludedResources{}, ccv3.Warnings{"get-space-warning"}, nil)

			apps = []resources.Application{
				{
					Name: "some-app-name-B",
					GUID: "some-app-guid-B",
				},
				{
					Name: "some-app-name-A",
					GUID: "some-app-guid-A",
				},
			}

		})

		Describe("org information", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationReturns(org, ccv3.Warnings{"get-org-warning"}, nil)
			})

			It("returns org name in the summary", func() {
				Expect(warnings).To(ConsistOf("get-org-warning", "get-space-warning"))
				Expect(spaceSummary.OrgName).To(Equal(org.Name))
			})

			When("getting org info fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationReturns(
						ccv3.Organization{},
						ccv3.Warnings{"get-org-warning"},
						errors.New("get-org-error"),
					)
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("get-org-error"))
				})
			})
		})

		Describe("space information", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv3.Space{{GUID: "some-space-guid", Name: "some-space"}},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-space-warning"},
					nil,
				)
			})

			It("returns space name in the summary", func() {
				Expect(warnings).To(ConsistOf("get-space-warning"))
				Expect(spaceSummary.Name).To(Equal("some-space"))
			})

			When("getting space info fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv3.Space{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get-space-warning"},
						errors.New("get-space-error"),
					)
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("get-space-error"))
				})
			})
		})

		Describe("app information", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(apps, ccv3.Warnings{"get-apps-warning"}, nil)
			})

			It("returns app names in the summary", func() {
				Expect(warnings).To(ConsistOf("get-apps-warning", "get-space-warning"))
				Expect(spaceSummary.AppNames).To(Equal([]string{apps[1].Name, apps[0].Name}))
			})

			When("getting app info fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]resources.Application{},
						ccv3.Warnings{"get-apps-warning"},
						errors.New("get-app-error"),
					)
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("get-app-error"))
				})
			})
		})

		Describe("service instance information", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns(
					[]ccv3.ServiceInstance{{Name: "instance-1"}, {Name: "instance-2"}},
					ccv3.Warnings{"get-services-warning"},
					nil,
				)
			})

			It("returns service instance names in the summary", func() {
				Expect(warnings).To(ConsistOf("get-services-warning", "get-space-warning"))
				Expect(spaceSummary.ServiceInstanceNames).To(Equal([]string{"instance-1", "instance-2"}))
			})

			When("getting service instance info fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstancesReturns(
						[]ccv3.ServiceInstance{},
						ccv3.Warnings{"get-services-warning"},
						errors.New("service-instance-error"),
					)
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("service-instance-error"))
				})
			})
		})

		Describe("isolation segment information", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceIsolationSegmentReturns(
					resources.Relationship{GUID: "iso-seg-guid"},
					ccv3.Warnings{"get-space-iso-seg-warning"},
					nil,
				)

				fakeCloudControllerClient.GetIsolationSegmentReturns(
					ccv3.IsolationSegment{GUID: "iso-seg-guid", Name: "some-iso-seg"},
					ccv3.Warnings{"get-iso-seg-warning"},
					nil,
				)
			})

			It("returns isolation segment name in the summary", func() {
				Expect(warnings).To(ConsistOf("get-space-iso-seg-warning", "get-iso-seg-warning", "get-space-warning"))
				Expect(spaceSummary.IsolationSegmentName).To(Equal("some-iso-seg"))
			})

			When("getting isolation segment info fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetIsolationSegmentReturns(
						ccv3.IsolationSegment{},
						ccv3.Warnings{"get-iso-seg-warning"},
						errors.New("iso-seg-error"),
					)
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("iso-seg-error"))
				})
			})
		})

		Describe("quota information", func() {
			BeforeEach(func() {
				ccv3Spaces = []ccv3.Space{
					{
						GUID: "some-space-guid",
						Name: "some-space-name",
						Relationships: resources.Relationships{
							constant.RelationshipTypeQuota: {
								GUID: "squota-guid",
							},
						},
					},
				}
				fakeCloudControllerClient.GetSpacesReturns(ccv3Spaces, ccv3.IncludedResources{}, ccv3.Warnings{"get-space-warning"}, nil)

				fakeCloudControllerClient.GetSpaceQuotaReturns(
					resources.SpaceQuota{Quota: resources.Quota{Name: "some-quota"}},
					ccv3.Warnings{"get-squota-warning"},
					nil,
				)
			})

			It("returns applied space quota name in the summary", func() {
				Expect(warnings).To(ConsistOf("get-squota-warning", "get-space-warning"))
				Expect(spaceSummary.QuotaName).To(Equal("some-quota"))
			})

			When("the space does not have a quota applied", func() {
				BeforeEach(func() {
					ccv3Spaces = []ccv3.Space{
						{
							GUID: "some-space-guid",
							Name: "some-space-name",
						},
					}
					fakeCloudControllerClient.GetSpacesReturns(ccv3Spaces, ccv3.IncludedResources{}, ccv3.Warnings{"get-space-warning"}, nil)
				})

				It("does not have a space quota name in the summary", func() {
					Expect(warnings).To(ConsistOf("get-space-warning"))
					Expect(spaceSummary.QuotaName).To(Equal(""))
				})
			})

			When("getting quota info fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceQuotaReturns(
						resources.SpaceQuota{},
						ccv3.Warnings{"get-squota-warning"},
						errors.New("space-quota-error"),
					)
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("space-quota-error"))
				})
			})
		})

		Describe("running security group information", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRunningSecurityGroupsReturns(
					[]resources.SecurityGroup{{Name: "run-group-1"}},
					ccv3.Warnings{"get-running-warning"},
					nil,
				)
			})

			It("returns running security group names in the summary", func() {
				Expect(warnings).To(ConsistOf("get-running-warning", "get-space-warning"))
				Expect(spaceSummary.RunningSecurityGroups).To(Equal([]resources.SecurityGroup{
					{Name: "run-group-1"},
				}))
			})

			When("getting running security group info fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRunningSecurityGroupsReturns(
						[]resources.SecurityGroup{},
						ccv3.Warnings{"get-running-warning"},
						errors.New("get-running-error"),
					)
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("get-running-error"))
				})
			})
		})

		Describe("staging security group information", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetStagingSecurityGroupsReturns(
					[]resources.SecurityGroup{{Name: "stag-group-1"}},
					ccv3.Warnings{"get-staging-warning"},
					nil,
				)
			})

			It("returns staging security group names in the summary", func() {
				Expect(warnings).To(ConsistOf("get-staging-warning", "get-space-warning"))
				Expect(spaceSummary.StagingSecurityGroups).To(Equal([]resources.SecurityGroup{
					{Name: "stag-group-1"},
				}))
			})

			When("getting staging security group info fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetStagingSecurityGroupsReturns(
						[]resources.SecurityGroup{},
						ccv3.Warnings{"get-staging-warning"},
						errors.New("get-staging-error"),
					)
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("get-staging-error"))
				})
			})
		})
	})
})
