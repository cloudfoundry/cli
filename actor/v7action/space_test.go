package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"

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
						[]ccv3.Space{
							{
								GUID: "some-space-guid",
								Name: spaceName,
							},
						},
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
							Relationships: ccv3.Relationships{
								constant.RelationshipTypeQuota: ccv3.Relationship{GUID: "some-space-quota-guid"},
							}}},
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
						Relationships: ccv3.Relationships{
							constant.RelationshipTypeQuota: ccv3.Relationship{GUID: "some-space-quota-guid"},
						},
					}))
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
		)

		JustBeforeEach(func() {
			spaceSummary, warnings, err = actor.GetSpaceSummaryByNameAndOrganization("space-name", "org-guid")
		})

		When("getting the organization succeeds", func() {
			var org ccv3.Organization

			BeforeEach(func() {
				org = ccv3.Organization{
					GUID: "some-org-guid",
					Name: "some-org-name",
				}

				w := ccv3.Warnings{"get-org-warning"}
				fakeCloudControllerClient.GetOrganizationReturns(org, w, nil)
			})

			When("getting the space succeeds", func() {
				var ccv3Spaces []ccv3.Space

				BeforeEach(func() {
					ccv3Spaces = []ccv3.Space{
						{
							GUID: "some-space-guid",
							Name: "some-space-name",
						},
					}

					w := ccv3.Warnings{"get-space-warning"}
					fakeCloudControllerClient.GetSpacesReturns(ccv3Spaces, w, nil)
				})

				When("getting the space applications succeeds", func() {
					var apps []ccv3.Application

					BeforeEach(func() {
						apps = []ccv3.Application{
							{
								Name: "some-app-name-B",
								GUID: "some-app-guid-B",
							},
							{
								Name: "some-app-name-A",
								GUID: "some-app-guid-A",
							},
						}

						w := ccv3.Warnings{"get-apps-warning"}
						fakeCloudControllerClient.GetApplicationsReturns(apps, w, nil)
					})

					When("getting the service instances succeeds", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetServiceInstancesReturns(
								[]ccv3.ServiceInstance{{Name: "instance-1"}, {Name: "instance-2"}},
								ccv3.Warnings{"get-services-warning"},
								nil,
							)
						})

						It("returns a populated space summary with app names sorted alphabetically", func() {
							expectedSpaceSummary := SpaceSummary{
								Name:    ccv3Spaces[0].Name,
								OrgName: org.Name,
								Space: Space{
									Name: ccv3Spaces[0].Name,
									GUID: ccv3Spaces[0].GUID,
								},
								AppNames:             []string{"some-app-name-A", "some-app-name-B"},
								ServiceInstanceNames: []string{"instance-1", "instance-2"},
							}

							Expect(err).ToNot(HaveOccurred())
							Expect(spaceSummary).To(Equal(expectedSpaceSummary))
						})

						It("propagates all warnings", func() {
							Expect(warnings).To(ConsistOf("get-apps-warning", "get-services-warning", "get-space-warning", "get-org-warning"))
						})
					})

					When("getting the service instances fails", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetServiceInstancesReturns(
								[]ccv3.ServiceInstance{},
								ccv3.Warnings{"get-services-warning"},
								errors.New("get-services-error"),
							)
						})

						It("returns warnings and error", func() {
							Expect(err).To(MatchError("get-services-error"))
							Expect(warnings).To(ConsistOf("get-apps-warning", "get-services-warning", "get-space-warning", "get-org-warning"))
						})
					})

					When("getting the space isolation segment returns no isolation segment relationship", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetSpaceIsolationSegmentReturns(
								ccv3.Relationship{},
								ccv3.Warnings{"get-space-iso-seg-warning"},
								nil,
							)
						})

						When("getting the org default isolation segment returns an isolation segment relationship", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentReturns(
									ccv3.Relationship{GUID: "org-default-iso-seg-guid"},
									ccv3.Warnings{"get-org-default-iso-seg-warning"},
									nil,
								)
								fakeCloudControllerClient.GetIsolationSegmentReturns(
									ccv3.IsolationSegment{GUID: "org-default-iso-seg-guid", Name: "org-default-iso-seg-name"},
									ccv3.Warnings{"get-iso-seg-warning"},
									nil,
								)
							})

							It("returns a populated space summary with app names sorted alphabetically", func() {
								expectedSpaceSummary := SpaceSummary{
									Name:    ccv3Spaces[0].Name,
									OrgName: org.Name,
									Space: Space{
										Name: ccv3Spaces[0].Name,
										GUID: ccv3Spaces[0].GUID,
									},
									AppNames:             []string{"some-app-name-A", "some-app-name-B"},
									ServiceInstanceNames: []string{},
									IsolationSegmentName: "org-default-iso-seg-name (org default)",
								}

								Expect(err).ToNot(HaveOccurred())
								Expect(spaceSummary).To(Equal(expectedSpaceSummary))
							})

							It("propagates all warnings", func() {
								Expect(warnings).To(ConsistOf("get-apps-warning", "get-space-warning", "get-org-warning", "get-space-iso-seg-warning", "get-org-default-iso-seg-warning", "get-iso-seg-warning"))
							})
						})

						When("getting the org default isolation segment returns no isolation segment relationship", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentReturns(
									ccv3.Relationship{},
									ccv3.Warnings{"get-org-default-iso-seg-warning"},
									nil,
								)
							})

							It("does not try to get the isolation segment", func() {
								Expect(fakeCloudControllerClient.GetIsolationSegmentCallCount()).To(Equal(0))
							})

							It("returns a populated space summary with app names sorted alphabetically", func() {
								expectedSpaceSummary := SpaceSummary{
									Name:    ccv3Spaces[0].Name,
									OrgName: org.Name,
									Space: Space{
										Name: ccv3Spaces[0].Name,
										GUID: ccv3Spaces[0].GUID,
									},
									AppNames:             []string{"some-app-name-A", "some-app-name-B"},
									ServiceInstanceNames: []string{},
								}

								Expect(err).ToNot(HaveOccurred())
								Expect(spaceSummary).To(Equal(expectedSpaceSummary))
							})

							It("propagates all warnings", func() {
								Expect(warnings).To(ConsistOf("get-apps-warning", "get-space-warning", "get-org-warning", "get-space-iso-seg-warning", "get-org-default-iso-seg-warning"))
							})
						})

						When("getting the org default isolation segment fails", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentReturns(
									ccv3.Relationship{},
									ccv3.Warnings{"get-org-default-iso-seg-warning"},
									errors.New("get-org-default-iso-seg-error"),
								)
							})

							It("returns all warnings and error", func() {
								Expect(err).To(MatchError("get-org-default-iso-seg-error"))
								Expect(warnings).To(ConsistOf("get-apps-warning", "get-space-warning", "get-org-warning", "get-space-iso-seg-warning", "get-org-default-iso-seg-warning"))
							})
						})
					})

					When("getting the space isolation segment returns an isolation segment relationship", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetSpaceIsolationSegmentReturns(
								ccv3.Relationship{GUID: "some-iso-seg-guid"},
								ccv3.Warnings{"get-space-iso-seg-warning"},
								nil,
							)
						})

						When("getting the isolation segment succeeds", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.GetIsolationSegmentReturns(
									ccv3.IsolationSegment{GUID: "some-iso-seg-guid", Name: "some-iso-seg-name"},
									ccv3.Warnings{"get-iso-seg-warning"},
									nil,
								)
							})

							It("returns a populated space summary with app names sorted alphabetically", func() {
								expectedSpaceSummary := SpaceSummary{
									Name:    ccv3Spaces[0].Name,
									OrgName: org.Name,
									Space: Space{
										Name: ccv3Spaces[0].Name,
										GUID: ccv3Spaces[0].GUID,
									},
									AppNames:             []string{"some-app-name-A", "some-app-name-B"},
									ServiceInstanceNames: []string{},
									IsolationSegmentName: "some-iso-seg-name",
								}

								Expect(err).ToNot(HaveOccurred())
								Expect(spaceSummary).To(Equal(expectedSpaceSummary))
							})

							It("propagates all warnings", func() {
								Expect(warnings).To(ConsistOf("get-apps-warning", "get-space-warning", "get-org-warning", "get-space-iso-seg-warning", "get-iso-seg-warning"))
							})
						})

						When("getting the isolation segment fails", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.GetIsolationSegmentReturns(
									ccv3.IsolationSegment{},
									ccv3.Warnings{"get-iso-seg-warning"},
									errors.New("get-iso-seg-error"),
								)
							})

							It("returns all warnings and the error", func() {
								Expect(err).To(MatchError("get-iso-seg-error"))
								Expect(warnings).To(ConsistOf("get-apps-warning", "get-space-warning", "get-org-warning", "get-space-iso-seg-warning", "get-iso-seg-warning"))
							})
						})
					})

					When("getting the space isolation segment fails", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetSpaceIsolationSegmentReturns(
								ccv3.Relationship{},
								ccv3.Warnings{"get-space-iso-seg-warning"},
								errors.New("get-space-iso-seg-error"),
							)
						})

						It("returns all warnings and the error", func() {
							Expect(err).To(MatchError("get-space-iso-seg-error"))
							Expect(warnings).To(ConsistOf("get-apps-warning", "get-space-warning", "get-org-warning", "get-space-iso-seg-warning"))
						})
					})

					When("there is no quota applied to the space", func() {
						It("requests never attempts to request the space quota", func() {
							Expect(fakeCloudControllerClient.GetSpaceQuotaCallCount()).To(Equal(0))
						})

						It("returns a space summary with an empty quota name", func() {
							Expect(spaceSummary.QuotaName).To(Equal(""))
						})

					})

					When("there is a quota applied to the space", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetSpacesReturns(
								[]ccv3.Space{
									{
										GUID: "some-space-guid",
										Name: "some-space-name",
										Relationships: ccv3.Relationships{
											constant.RelationshipTypeQuota: ccv3.Relationship{GUID: "some-space-quota-guid"},
										},
									},
								},
								ccv3.Warnings{"get-space-warning"}, nil)
						})

						It("requests the space quota by the applied quota GUID", func() {
							Expect(fakeCloudControllerClient.GetSpaceQuotaCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetSpaceQuotaArgsForCall(0)).To(Equal("some-space-quota-guid"))
						})

						When("getting the space quota succeeds", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.GetSpaceQuotaReturns(
									resources.SpaceQuota{
										Quota: resources.Quota{
											Name: "some-space-quota",
											GUID: "some-space-quota-guid",
										},
									},
									ccv3.Warnings{"get-space-quota-warning"},
									nil,
								)
							})

							It("returns the applied quota name on the space summary", func() {
								Expect(spaceSummary.QuotaName).To(Equal("some-space-quota"))
							})
						})
						When("getting the space quota fails", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.GetSpaceQuotaReturns(
									resources.SpaceQuota{
										Quota: resources.Quota{
											Name: "some-space-quota",
											GUID: "some-space-quota-guid",
										},
									},
									ccv3.Warnings{"get-space-quota-warning"},
									errors.New("get-space-quota-error"),
								)
							})

							It("returns all warnings and an error", func() {
								Expect(err).To(MatchError("get-space-quota-error"))
								Expect(warnings).To(ConsistOf("get-org-warning", "get-space-warning", "get-apps-warning", "get-space-quota-warning"))

							})
						})
					})
				})

				When("getting the space applications fails", func() {
					BeforeEach(func() {
						e := errors.New("get-apps-error")
						w := ccv3.Warnings{"get-apps-warning"}
						fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{}, w, e)
					})

					It("returns all warnings and an error", func() {
						Expect(err).To(MatchError("get-apps-error"))
						Expect(warnings).To(ConsistOf("get-apps-warning", "get-space-warning", "get-org-warning"))
					})
				})
			})

			When("getting the space fails", func() {
				BeforeEach(func() {
					e := errors.New("get-space-error")
					w := ccv3.Warnings{"get-space-warning"}
					fakeCloudControllerClient.GetSpacesReturns([]ccv3.Space{}, w, e)
				})

				It("returns all warnings and an error", func() {
					Expect(err).To(MatchError("get-space-error"))
					Expect(warnings).To(ConsistOf("get-space-warning", "get-org-warning"))
				})
			})
		})

		When("getting the organization fails", func() {
			BeforeEach(func() {
				e := errors.New("get-org-error")
				w := ccv3.Warnings{"get-org-warning"}
				fakeCloudControllerClient.GetOrganizationReturns(ccv3.Organization{}, w, e)
			})

			It("returns warnings and an error", func() {
				Expect(err).To(MatchError("get-org-error"))
				Expect(warnings).To(ConsistOf("get-org-warning"))
			})
		})
	})
})
