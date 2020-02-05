package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	uaaconst "code.cloudfoundry.org/cli/api/uaa/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		fakeConfig                *v2actionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v2actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, nil, fakeConfig)
	})

	Describe("CreateSpace", func() {
		var (
			quotaName string

			space      Space
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			quotaName = ""
		})

		JustBeforeEach(func() {
			space, warnings, executeErr = actor.CreateSpace("some-space", "some-org-name", quotaName)
		})

		When("the org is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{
						"warning-1",
						"warning-2",
					},
					actionerror.OrganizationNotFoundError{Name: "some-org-name"},
				)
			})

			It("returns an OrganizationNotFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org-name"}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("the org is found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv2.Organization{
					ccv2.Organization{GUID: "some-org-guid", Name: "some-org-name"}},
					ccv2.Warnings{},
					nil,
				)
			})

			When("the specified quota is not found", func() {
				BeforeEach(func() {
					quotaName = "some-space-quota"
					fakeCloudControllerClient.GetSpaceQuotasReturns(
						[]ccv2.SpaceQuota{},
						ccv2.Warnings{
							"get-quota-warning-1",
							"get-quota-warning-2",
						},
						nil,
					)
				})

				It("returns a QuotaNotFoundForNameError", func() {
					Expect(executeErr).To(MatchError(actionerror.QuotaNotFoundForNameError{Name: quotaName}))
					Expect(warnings).To(ContainElement("get-quota-warning-1"))
					Expect(warnings).To(ContainElement("get-quota-warning-2"))
				})

				It("does not create the space", func() {
					Expect(fakeCloudControllerClient.CreateSpaceCallCount()).To(Equal(0))
				})
			})

			// When("the specified quota is found")

			// When("no quota is specified")

			When("creating the space succeeds", func() {
				var expectedSpace Space
				BeforeEach(func() {
					expectedSpace = Space{
						GUID: "some-space-guid",
						Name: "some-space",
					}

					fakeCloudControllerClient.CreateSpaceReturns(
						ccv2.Space{
							GUID: "some-space-guid",
							Name: "some-space",
						},
						ccv2.Warnings{"create-space-warning-1", "create-space-warning-2"},
						nil)
				})

				It("should return the space and all its warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					spaceName, orgGuid := fakeCloudControllerClient.CreateSpaceArgsForCall(0)
					Expect(spaceName).To(Equal("some-space"))
					Expect(orgGuid).To(Equal("some-org-guid"))

					Expect(space).To(Equal(expectedSpace))
					Expect(warnings).To(ConsistOf("create-space-warning-1", "create-space-warning-2"))
				})

				When("the quota is found", func() {
					BeforeEach(func() {
						spaceQuotaGuid := "space-quota-guid"
						quotaName = "some-quota"
						fakeCloudControllerClient.GetSpaceQuotasReturns(
							[]ccv2.SpaceQuota{
								ccv2.SpaceQuota{Name: quotaName, GUID: spaceQuotaGuid},
							},
							ccv2.Warnings{
								"get-quota-warning-1",
								"get-quota-warning-2",
							},
							nil,
						)
					})

					It("sets the quota on the space", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ContainElement("get-quota-warning-1"))
						Expect(warnings).To(ContainElement("get-quota-warning-2"))
						Expect(fakeCloudControllerClient.SetSpaceQuotaCallCount()).To(Equal(1))
					})
				})
			})

			When("the space name is taken", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateSpaceReturns(ccv2.Space{}, ccv2.Warnings{"some-warning"}, ccerror.SpaceNameTakenError{
						Message: "nice try",
					})
				})

				It("returns a SpaceNameTakenError", func() {
					Expect(executeErr).To(MatchError(actionerror.SpaceNameTakenError{Name: "some-space"}))
					Expect(warnings).To(ConsistOf("some-warning"))
				})
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

		When("the org is found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{Name: "some-org", GUID: "some-org-guid"}},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			When("the space is not found", func() {
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

			When("the space is found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{{GUID: "some-space-guid"}},
						ccv2.Warnings{"warning-3", "warning-4"},
						nil,
					)
				})

				When("the delete returns an error", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("some delete space error")
						fakeCloudControllerClient.DeleteSpaceJobReturns(
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

				When("the delete returns a job", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.DeleteSpaceJobReturns(
							ccv2.Job{GUID: "some-job-guid"},
							ccv2.Warnings{"warning-5", "warning-6"},
							nil,
						)
					})

					When("polling errors", func() {
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

					When("the job is successful", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.PollJobReturns(ccv2.Warnings{"warning-7", "warning-8"}, nil)
						})

						It("returns warnings and no error", func() {
							Expect(err).ToNot(HaveOccurred())
							Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6", "warning-7", "warning-8"))

							Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(Equal([]ccv2.Filter{{
								Type:     constant.NameFilter,
								Operator: constant.EqualOperator,
								Values:   []string{"some-org"},
							}}))

							Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal([]ccv2.Filter{{
								Type:     constant.NameFilter,
								Operator: constant.EqualOperator,
								Values:   []string{"some-space"},
							},
								{
									Type:     constant.OrganizationGUIDFilter,
									Operator: constant.EqualOperator,
									Values:   []string{"some-org-guid"},
								},
							}))

							Expect(fakeCloudControllerClient.DeleteSpaceJobCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.DeleteSpaceJobArgsForCall(0)).To(Equal("some-space-guid"))

							Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.PollJobArgsForCall(0)).To(Equal(ccv2.Job{GUID: "some-job-guid"}))
						})
					})
				})
			})
		})
	})

	Describe("GetOrganizationSpaces", func() {
		When("there are spaces in the org", func() {
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
					[]ccv2.Filter{
						{
							Type:     constant.OrganizationGUIDFilter,
							Operator: constant.EqualOperator,
							Values:   []string{"some-org-guid"},
						},
					}))
			})
		})

		When("an error is encountered", func() {
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
		var (
			space      Space
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			space, warnings, executeErr = actor.GetSpaceByOrganizationAndName("some-org-guid", "some-space")
		})

		When("the space exists", func() {
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
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(space).To(Equal(Space{
					GUID:                     "some-space-guid",
					Name:                     "some-space",
					AllowSSH:                 true,
					SpaceQuotaDefinitionGUID: "some-space-quota-guid",
				}))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(
					[]ccv2.Filter{
						{
							Type:     constant.OrganizationGUIDFilter,
							Operator: constant.EqualOperator,
							Values:   []string{"some-org-guid"},
						},
						{
							Type:     constant.NameFilter,
							Operator: constant.EqualOperator,
							Values:   []string{"some-space"},
						},
					}))
			})
		})

		When("an error is encountered", func() {
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
				Expect(executeErr).To(MatchError(returnedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("the space does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{},
					nil,
					nil,
				)
			})

			It("returns SpaceNotFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{
					Name: "some-space",
				}))
			})
		})

		When("multiple spaces exists", func() {
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
				Expect(executeErr).To(MatchError(actionerror.MultipleSpacesFoundError{
					OrgGUID: "some-org-guid",
					Name:    "some-space",
				}))
			})
		})
	})

	Describe("GrantSpaceManagerByUsername", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.GrantSpaceManagerByUsername("some-org-guid", "some-space-guid", "some-username")
		})

		When("acting as a user", func() {
			When("ensuring the user is a member of the org succeeds", func() {
				When("the cloud controller returns with success", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.UpdateSpaceManagerByUsernameReturns(
							ccv2.Warnings{"warning-1", "warning-2"},
							nil,
						)
						fakeCloudControllerClient.UpdateOrganizationUserByUsernameReturns(ccv2.Warnings{"org-user-warning"}, nil)
					})

					It("returns all the warnings", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(warnings).To(ConsistOf("warning-1", "warning-2", "org-user-warning"))
					})

					It("makes the user a member of the org", func() {
						Expect(fakeCloudControllerClient.UpdateOrganizationUserByUsernameCallCount()).To(Equal(1))
						orgGUID, username := fakeCloudControllerClient.UpdateOrganizationUserByUsernameArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))
						Expect(username).To(Equal("some-username"))
					})

					It("makes the user a space manager in the requested space", func() {
						Expect(fakeCloudControllerClient.UpdateSpaceManagerByUsernameCallCount()).To(Equal(1))
					})
				})
			})

			When("ensuring the user is a member of the org fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateOrganizationUserByUsernameReturns(
						ccv2.Warnings{"some org user warning"},
						errors.New("could not add user to org"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ConsistOf("some org user warning"))
					Expect(executeErr).To(MatchError("could not add user to org"))
				})

				It("does not make the user a space manager", func() {
					Expect(fakeCloudControllerClient.UpdateSpaceManagerByUsernameCallCount()).To(Equal(0))
				})
			})

			When("making the user a space manager errors", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = errors.New("cc-grant-space-manager-error")
					fakeCloudControllerClient.UpdateSpaceManagerByUsernameReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						returnedErr,
					)
				})

				It("returns the error and all the warnings", func() {
					Expect(executeErr).To(MatchError(returnedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			It("does not treat the username as a guid", func() {
				Expect(fakeCloudControllerClient.UpdateOrganizationUserCallCount()).To(Equal(0))
				Expect(fakeCloudControllerClient.UpdateSpaceManagerCallCount()).To(Equal(0))
			})
		})

		When("acting as a client", func() {
			BeforeEach(func() {
				fakeConfig.UAAGrantTypeReturns(string(uaaconst.GrantTypeClientCredentials))
			})

			When("ensuring the client is a member of the org succeeds", func() {
				When("the cloud controller returns with success", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.UpdateOrganizationUserReturns(ccv2.Warnings{"org-user-warning"}, nil)
						fakeCloudControllerClient.UpdateSpaceManagerReturns(
							ccv2.Warnings{"warning-1", "warning-2"},
							nil,
						)
					})

					It("returns all the warnings", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(warnings).To(ConsistOf("warning-1", "warning-2", "org-user-warning"))
					})

					It("makes the client a member of the org", func() {
						Expect(fakeCloudControllerClient.UpdateOrganizationUserCallCount()).To(Equal(1))
						orgGUID, username := fakeCloudControllerClient.UpdateOrganizationUserArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))
						Expect(username).To(Equal("some-username"))
					})

					It("makes the user a space manager for the requested space", func() {
						Expect(fakeCloudControllerClient.UpdateSpaceManagerCallCount()).To(Equal(1))
						spaceGUID, clientID := fakeCloudControllerClient.UpdateSpaceManagerArgsForCall(0)
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(clientID).To(Equal("some-username"))
					})
				})
			})

			When("ensuring the client is a member of the org fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateOrganizationUserReturns(
						ccv2.Warnings{"some org user warning"},
						errors.New("ya done goofed"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ConsistOf("some org user warning"))
					Expect(executeErr).To(MatchError("ya done goofed"))
				})

				It("does not make the user a space manager", func() {
					Expect(fakeCloudControllerClient.UpdateSpaceManagerCallCount()).To(Equal(0))
				})
			})

			When("making the client a space manager errors", func() {
				var returnedErr error

				BeforeEach(func() {
					fakeCloudControllerClient.UpdateOrganizationUserReturns(
						ccv2.Warnings{"some org user warning"},
						nil,
					)

					returnedErr = errors.New("cc-grant-space-manager-error")
					fakeCloudControllerClient.UpdateSpaceManagerReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						returnedErr,
					)
				})

				It("returns the error and all the warnings", func() {
					Expect(executeErr).To(MatchError(returnedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "some org user warning"))
				})
			})

			It("does not treat the client ID as a username", func() {
				Expect(fakeCloudControllerClient.UpdateOrganizationUserByUsernameCallCount()).To(Equal(0))
				Expect(fakeCloudControllerClient.UpdateSpaceManagerByUsernameCallCount()).To(Equal(0))
			})
		})
	})

	Describe("GrantSpaceDeveloperByUsername", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.GrantSpaceDeveloperByUsername("some-space-guid", "some-username")
		})

		When("acting as a user", func() {
			It("attempts to add the user to the space developers list", func() {
				Expect(fakeCloudControllerClient.UpdateSpaceDeveloperByUsernameCallCount()).To(Equal(1))
				spaceGUID, username := fakeCloudControllerClient.UpdateSpaceDeveloperByUsernameArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(username).To(Equal("some-username"))
			})

			When("the request is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateSpaceDeveloperByUsernameReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("returns all warnings", func() {
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			When("the request fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateSpaceDeveloperByUsernameReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						errors.New("the request failed"),
					)
				})

				It("return the error", func() {
					Expect(executeErr).To(MatchError("the request failed"))
				})

				It("returns all warnings", func() {
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})

		When("acting as a client", func() {
			BeforeEach(func() {
				fakeConfig.UAAGrantTypeReturns(string(uaaconst.GrantTypeClientCredentials))
			})

			It("attempts to add the client to the space developers list", func() {
				Expect(fakeCloudControllerClient.UpdateSpaceDeveloperCallCount()).To(Equal(1))
				spaceGUID, username := fakeCloudControllerClient.UpdateSpaceDeveloperArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(username).To(Equal("some-username"))
			})

			It("does not treat the client ID as a username", func() {
				Expect(fakeCloudControllerClient.UpdateSpaceDeveloperByUsernameCallCount()).To(Equal(0))
			})

			When("making the user a space developer succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateSpaceDeveloperReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("returns all the warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			When("making the user a space developer fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateSpaceDeveloperReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						errors.New("boom!"),
					)
				})

				It("returns all the warnings", func() {
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("boom!"))
				})
			})
		})
	})
})
