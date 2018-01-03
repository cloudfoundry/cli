package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space", func() {
	Describe("SpaceNotFoundError#Error", func() {
		Context("when the name is specified", func() {
			It("returns an error message with the name of the missing space", func() {
				err := actionerror.SpaceNotFoundError{
					Name: "some-space",
				}
				Expect(err.Error()).To(Equal("Space 'some-space' not found."))
			})
		})

		Context("when the name is not specified, but the GUID is specified", func() {
			It("returns an error message with the GUID of the missing space", func() {
				err := actionerror.SpaceNotFoundError{
					GUID: "some-space-guid",
				}
				Expect(err.Error()).To(Equal("Space with GUID 'some-space-guid' not found."))
			})
		})

		Context("when neither the name nor the GUID is specified", func() {
			It("returns a generic error message for the missing space", func() {
				err := actionerror.SpaceNotFoundError{}
				Expect(err.Error()).To(Equal("Space '' not found."))
			})
		})
	})

	Describe("Actions", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil)
		})

		Describe("DeleteSpaceByNameAndOrganizationName", func() {
			var (
				warnings Warnings
				err      error
			)

			JustBeforeEach(func() {
				warnings, err = actor.DeleteSpaceByNameAndOrganizationName("some-space", "some-org")
			})

			Context("when the org is not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationsReturns(
						[]ccv2.Organization{},
						ccv2.Warnings{
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

			Context("when the org is found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationsReturns(
						[]ccv2.Organization{{Name: "some-org", GUID: "some-org-guid"}},
						ccv2.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				Context("when the space is not found", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetSpacesReturns(
							[]ccv2.Space{},
							ccv2.Warnings{"warning-3", "warning-4"},
							nil,
						)
					})

					It("returns an SpaceNotFoundError", func() {
						Expect(err).To(MatchError(actionerror.SpaceNotFoundError{Name: "some-space"}))
						Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
					})
				})

				Context("when the space is found", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetSpacesReturns(
							[]ccv2.Space{{GUID: "some-space-guid"}},
							ccv2.Warnings{"warning-3", "warning-4"},
							nil,
						)
					})

					Context("when the delete returns an error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("some delete space error")
							fakeCloudControllerClient.DeleteSpaceReturns(
								ccv2.Job{},
								ccv2.Warnings{"warning-5", "warning-6"},
								expectedErr,
							)
						})

						It("returns the error", func() {
							Expect(err).To(Equal(expectedErr))
							Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6"))
						})
					})

					Context("when the delete returns a job", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.DeleteSpaceReturns(
								ccv2.Job{GUID: "some-job-guid"},
								ccv2.Warnings{"warning-5", "warning-6"},
								nil,
							)
						})

						Context("when polling errors", func() {
							var expectedErr error

							BeforeEach(func() {
								expectedErr = errors.New("Never expected, by anyone")
								fakeCloudControllerClient.PollJobReturns(ccv2.Warnings{"warning-7", "warning-8"}, expectedErr)
							})

							It("returns the error", func() {
								Expect(err).To(Equal(expectedErr))
								Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6", "warning-7", "warning-8"))
							})
						})

						Context("when the job is successful", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.PollJobReturns(ccv2.Warnings{"warning-7", "warning-8"}, nil)
							})

							It("returns warnings and no error", func() {
								Expect(err).ToNot(HaveOccurred())
								Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6", "warning-7", "warning-8"))

								Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
								Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(Equal([]ccv2.QQuery{{
									Filter:   ccv2.NameFilter,
									Operator: ccv2.EqualOperator,
									Values:   []string{"some-org"},
								}}))

								Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
								Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal([]ccv2.QQuery{{
									Filter:   ccv2.NameFilter,
									Operator: ccv2.EqualOperator,
									Values:   []string{"some-space"},
								},
									{
										Filter:   ccv2.OrganizationGUIDFilter,
										Operator: ccv2.EqualOperator,
										Values:   []string{"some-org-guid"},
									},
								}))

								Expect(fakeCloudControllerClient.DeleteSpaceCallCount()).To(Equal(1))
								Expect(fakeCloudControllerClient.DeleteSpaceArgsForCall(0)).To(Equal("some-space-guid"))

								Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
								Expect(fakeCloudControllerClient.PollJobArgsForCall(0)).To(Equal(ccv2.Job{GUID: "some-job-guid"}))
							})
						})
					})
				})
			})
		})

		Describe("GetOrganizationSpaces", func() {
			Context("when there are spaces in the org", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{
							{
								GUID:                     "space-1-guid",
								Name:                     "space-1",
								AllowSSH:                 true,
								SpaceQuotaDefinitionGUID: "some-space-quota-guid",
							},
							{
								GUID:     "space-2-guid",
								Name:     "space-2",
								AllowSSH: false,
							},
						},
						ccv2.Warnings{"warning-1", "warning-2"},
						nil)
				})

				It("returns all spaces and all warnings", func() {
					spaces, warnings, err := actor.GetOrganizationSpaces("some-org-guid")

					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(spaces).To(Equal(
						[]Space{
							{
								GUID:                     "space-1-guid",
								Name:                     "space-1",
								AllowSSH:                 true,
								SpaceQuotaDefinitionGUID: "some-space-quota-guid",
							},
							{
								GUID:     "space-2-guid",
								Name:     "space-2",
								AllowSSH: false,
							},
						}))

					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal(
						[]ccv2.QQuery{
							{
								Filter:   ccv2.OrganizationGUIDFilter,
								Operator: ccv2.EqualOperator,
								Values:   []string{"some-org-guid"},
							},
						}))
				})
			})

			Context("when an error is encountered", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = errors.New("cc-get-spaces-error")
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{},
						ccv2.Warnings{"warning-1", "warning-2"},
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

		Describe("GetSpaceByOrganizationAndName", func() {
			Context("when the space exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{
							{
								GUID:                     "some-space-guid",
								Name:                     "some-space",
								AllowSSH:                 true,
								SpaceQuotaDefinitionGUID: "some-space-quota-guid",
							},
						},
						ccv2.Warnings{"warning-1", "warning-2"},
						nil)
				})

				It("returns the space and all warnings", func() {
					space, warnings, err := actor.GetSpaceByOrganizationAndName("some-org-guid", "some-space")

					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(space).To(Equal(Space{
						GUID:                     "some-space-guid",
						Name:                     "some-space",
						AllowSSH:                 true,
						SpaceQuotaDefinitionGUID: "some-space-quota-guid",
					}))

					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(
						[]ccv2.QQuery{
							{
								Filter:   ccv2.OrganizationGUIDFilter,
								Operator: ccv2.EqualOperator,
								Values:   []string{"some-org-guid"},
							},
							{
								Filter:   ccv2.NameFilter,
								Operator: ccv2.EqualOperator,
								Values:   []string{"some-space"},
							},
						}))
				})
			})

			Context("when an error is encountered", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = errors.New("cc-get-spaces-error")
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{},
						ccv2.Warnings{"warning-1", "warning-2"},
						returnedErr,
					)
				})

				It("return the error and all warnings", func() {
					_, warnings, err := actor.GetSpaceByOrganizationAndName("some-org-guid", "some-space")

					Expect(err).To(MatchError(returnedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			Context("when the space does not exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{},
						nil,
						nil,
					)
				})

				It("returns SpaceNotFoundError", func() {
					_, _, err := actor.GetSpaceByOrganizationAndName("some-org-guid", "some-space")

					Expect(err).To(MatchError(actionerror.SpaceNotFoundError{
						Name: "some-space",
					}))
				})
			})

			Context("when multiple spaces exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{
							{
								GUID:     "some-space-guid",
								Name:     "some-space",
								AllowSSH: true,
							},
							{
								GUID:     "another-space-guid",
								Name:     "another-space",
								AllowSSH: true,
							},
						},
						nil,
						nil,
					)
				})

				It("returns MultipleSpacesFoundError", func() {
					_, _, err := actor.GetSpaceByOrganizationAndName("some-org-guid", "some-space")

					Expect(err).To(MatchError(actionerror.MultipleSpacesFoundError{
						OrgGUID: "some-org-guid",
						Name:    "some-space",
					}))
				})
			})
		})
	})
})
