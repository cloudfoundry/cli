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
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Org Actions", func() {
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

	Describe("GetOrganization", func() {
		var (
			org      Organization
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			org, warnings, err = actor.GetOrganization("some-org-guid")
		})

		When("the org exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationReturns(
					ccv2.Organization{
						GUID:                "some-org-guid",
						Name:                "some-org",
						QuotaDefinitionGUID: "some-quota-definition-guid",
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("returns the org and all warnings", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(org.GUID).To(Equal("some-org-guid"))
				Expect(org.Name).To(Equal("some-org"))
				Expect(org.QuotaDefinitionGUID).To(Equal("some-quota-definition-guid"))

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetOrganizationCallCount()).To(Equal(1))
				guid := fakeCloudControllerClient.GetOrganizationArgsForCall(0)
				Expect(guid).To(Equal("some-org-guid"))
			})
		})

		When("the org does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationReturns(
					ccv2.Organization{},
					ccv2.Warnings{"warning-1", "warning-2"},
					ccerror.ResourceNotFoundError{},
				)
			})

			It("returns warnings and OrganizationNotFoundError", func() {
				Expect(err).To(MatchError(actionerror.OrganizationNotFoundError{GUID: "some-org-guid"}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some get org error")
				fakeCloudControllerClient.GetOrganizationReturns(
					ccv2.Organization{},
					ccv2.Warnings{"warning-1", "warning-2"},
					expectedErr,
				)
			})

			It("returns warnings and the error", func() {
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetOrganizationByName", func() {
		var (
			org      Organization
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			org, warnings, err = actor.GetOrganizationByName("some-org")
		})

		When("the org exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{
						{GUID: "some-org-guid"},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("returns the org and all warnings", func() {
				Expect(org.GUID).To(Equal("some-org-guid"))
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				filters := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(filters).To(Equal(
					[]ccv2.Filter{{
						Type:     constant.NameFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"some-org"},
					}}))
			})
		})

		When("the org does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					nil,
					nil,
				)
			})

			It("returns OrganizationNotFoundError", func() {
				Expect(err).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))
			})
		})

		When("multiple orgs exist with the same name", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{
						{GUID: "org-1-guid"},
						{GUID: "org-2-guid"},
					},
					nil,
					nil,
				)
			})

			It("returns MultipleOrganizationsFoundError", func() {
				Expect(err).To(MatchError("Organization name 'some-org' matches multiple GUIDs: org-1-guid, org-2-guid"))
			})
		})

		When("an error is encountered", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("get-orgs-error")
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{
						"warning-1",
						"warning-2",
					},
					returnedErr,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(returnedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GrantOrgManagerByUsername", func() {
		var (
			guid     string
			username string
			warnings Warnings
			err      error
		)
		JustBeforeEach(func() {
			warnings, err = actor.GrantOrgManagerByUsername(guid, username)
		})

		When("acting as a user", func() {
			When("making the user an org manager succeeds", func() {
				BeforeEach(func() {
					guid = "some-guid"
					username = "some-user"

					fakeCloudControllerClient.UpdateOrganizationManagerByUsernameReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("returns warnings", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeCloudControllerClient.UpdateOrganizationManagerByUsernameCallCount()).To(Equal(1))
					orgGuid, user := fakeCloudControllerClient.UpdateOrganizationManagerByUsernameArgsForCall(0)
					Expect(orgGuid).To(Equal("some-guid"))
					Expect(user).To(Equal("some-user"))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			When("making the user an org manager fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateOrganizationManagerByUsernameReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						errors.New("some-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(err).To(MatchError("some-error"))
				})
			})
		})

		When("acting as a client", func() {
			BeforeEach(func() {
				fakeConfig.UAAGrantTypeReturns(string(uaaconst.GrantTypeClientCredentials))
			})

			When("making the client an org manager succeeds", func() {
				BeforeEach(func() {
					guid = "some-guid"
					username = "some-client-id"

					fakeCloudControllerClient.UpdateOrganizationManagerReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("returns warnings", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeCloudControllerClient.UpdateOrganizationManagerCallCount()).To(Equal(1))
					orgGuid, user := fakeCloudControllerClient.UpdateOrganizationManagerArgsForCall(0)
					Expect(orgGuid).To(Equal("some-guid"))
					Expect(user).To(Equal("some-client-id"))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			When("making the client an org manager fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateOrganizationManagerReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						errors.New("some-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(err).To(MatchError("some-error"))
				})
			})
		})
	})

	Describe("CreateOrganization", func() {
		var (
			quotaName string

			org        Organization
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			org, warnings, executeErr = actor.CreateOrganization("some-org", quotaName)
		})

		When("a quota is not specified", func() {
			BeforeEach(func() {
				quotaName = ""
			})

			When("the organization is created successfully", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateOrganizationReturns(
						ccv2.Organization{
							GUID: "some-org-guid",
							Name: "some-org",
						},
						ccv2.Warnings{"warning-1", "warning-2"},
						nil)
				})

				It("returns the org and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(org.GUID).To(Equal("some-org-guid"))
					Expect(org.Name).To(Equal("some-org"))

					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.CreateOrganizationCallCount()).To(Equal(1))
					orgName, quotaGUID := fakeCloudControllerClient.CreateOrganizationArgsForCall(0)
					Expect(quotaGUID).To(BeEmpty())
					Expect(orgName).To(Equal("some-org"))
				})
			})

			When("creating the organzation fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateOrganizationReturns(
						ccv2.Organization{},
						ccv2.Warnings{"warning-1", "warning-2"},
						errors.New("couldn't make it"))
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(MatchError("couldn't make it"))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(fakeCloudControllerClient.CreateOrganizationCallCount()).To(Equal(1))
				})

				When("the organization name already exists", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.CreateOrganizationReturns(
							ccv2.Organization{},
							ccv2.Warnings{"warning-1", "warning-2"},
							ccerror.OrganizationNameTakenError{Message: "name is taken"},
						)
					})

					It("wraps the error in an action error", func() {
						Expect(executeErr).To(MatchError(actionerror.OrganizationNameTakenError{Name: "some-org"}))
						Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
						Expect(fakeCloudControllerClient.CreateOrganizationCallCount()).To(Equal(1))
					})
				})
			})
		})

		When("a quota name is specified", func() {
			BeforeEach(func() {
				quotaName = "some-quota"
			})

			When("the fetching the quota succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationQuotasReturns(
						[]ccv2.OrganizationQuota{
							{
								GUID: "some-quota-definition-guid",
								Name: "some-quota",
							},
						},
						ccv2.Warnings{"quota-warning-1", "quota-warning-2"},
						nil)
				})

				When("creating the org succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.CreateOrganizationReturns(
							ccv2.Organization{
								GUID:                "some-org-guid",
								Name:                "some-org",
								QuotaDefinitionGUID: "some-quota-definition-guid",
							},
							ccv2.Warnings{"warning-1", "warning-2"},
							nil)
					})

					It("includes that quota's guid when creating the org", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("quota-warning-1", "quota-warning-2", "warning-1", "warning-2"))
						Expect(org).To(MatchFields(IgnoreExtras, Fields{
							"GUID":                Equal("some-org-guid"),
							"Name":                Equal("some-org"),
							"QuotaDefinitionGUID": Equal("some-quota-definition-guid"),
						}))

						Expect(fakeCloudControllerClient.CreateOrganizationCallCount()).To(Equal(1))
						orgName, quotaGUID := fakeCloudControllerClient.CreateOrganizationArgsForCall(0)
						Expect(quotaGUID).To(Equal("some-quota-definition-guid"))
						Expect(orgName).To(Equal("some-org"))
					})
				})

				When("creating the org fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.CreateOrganizationReturns(
							ccv2.Organization{},
							ccv2.Warnings{"warning-1", "warning-2"},
							errors.New("couldn't make it"))
					})

					It("returns the error and warnings", func() {
						Expect(executeErr).To(MatchError("couldn't make it"))
						Expect(warnings).To(ConsistOf("quota-warning-1", "quota-warning-2", "warning-1", "warning-2"))

						Expect(fakeCloudControllerClient.CreateOrganizationCallCount()).To(Equal(1))
					})
				})
			})

			When("fetching the quota fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationQuotasReturns(
						nil,
						ccv2.Warnings{"quota-warning-1", "quota-warning-2"},
						errors.New("no quota found"))
				})

				It("returns warnings and the error, and does not try to create the org", func() {
					Expect(executeErr).To(MatchError("no quota found"))
					Expect(warnings).To(ConsistOf("quota-warning-1", "quota-warning-2"))

					Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.CreateOrganizationCallCount()).To(Equal(0))
				})
			})
		})
	})

	Describe("DeleteOrganization", func() {
		var (
			warnings     Warnings
			deleteOrgErr error
			job          ccv2.Job
		)

		JustBeforeEach(func() {
			warnings, deleteOrgErr = actor.DeleteOrganization("some-org")
		})

		Context("the organization is deleted successfully", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv2.Organization{
					{GUID: "some-org-guid"},
				}, ccv2.Warnings{"get-org-warning"}, nil)

				job = ccv2.Job{
					GUID:   "some-job-guid",
					Status: constant.JobStatusFinished,
				}

				fakeCloudControllerClient.DeleteOrganizationJobReturns(
					job, ccv2.Warnings{"delete-org-warning"}, nil)

				fakeCloudControllerClient.PollJobReturns(ccv2.Warnings{"polling-warnings"}, nil)
			})

			It("returns warnings and deletes the org", func() {
				Expect(warnings).To(ConsistOf("get-org-warning", "delete-org-warning", "polling-warnings"))
				Expect(deleteOrgErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				filters := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(filters).To(Equal(
					[]ccv2.Filter{{
						Type:     constant.NameFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"some-org"},
					}}))

				Expect(fakeCloudControllerClient.DeleteOrganizationJobCallCount()).To(Equal(1))
				orgGuid := fakeCloudControllerClient.DeleteOrganizationJobArgsForCall(0)
				Expect(orgGuid).To(Equal("some-org-guid"))

				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				job := fakeCloudControllerClient.PollJobArgsForCall(0)
				Expect(job.GUID).To(Equal("some-job-guid"))
			})
		})

		When("getting the org returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{
						"get-org-warning",
					},
					nil,
				)
			})

			It("returns an error and all warnings", func() {
				Expect(warnings).To(ConsistOf("get-org-warning"))
				Expect(deleteOrgErr).To(MatchError(actionerror.OrganizationNotFoundError{
					Name: "some-org",
				}))
			})
		})

		When("the delete returns an error", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("delete-org-error")

				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{GUID: "org-1-guid"}},
					ccv2.Warnings{
						"get-org-warning",
					},
					nil,
				)

				fakeCloudControllerClient.DeleteOrganizationJobReturns(
					ccv2.Job{},
					ccv2.Warnings{"delete-org-warning"},
					returnedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(deleteOrgErr).To(MatchError(returnedErr))
				Expect(warnings).To(ConsistOf("get-org-warning", "delete-org-warning"))
			})
		})

		When("the job polling has an error", func() {
			var expectedErr error
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv2.Organization{
					{GUID: "some-org-guid"},
				}, ccv2.Warnings{"get-org-warning"}, nil)

				fakeCloudControllerClient.DeleteOrganizationJobReturns(
					ccv2.Job{}, ccv2.Warnings{"delete-org-warning"}, nil)

				expectedErr = errors.New("Never expected, by anyone")
				fakeCloudControllerClient.PollJobReturns(ccv2.Warnings{"polling-warnings"}, expectedErr)
			})

			It("returns the error from job polling", func() {
				Expect(warnings).To(ConsistOf("get-org-warning", "delete-org-warning", "polling-warnings"))
				Expect(deleteOrgErr).To(MatchError(expectedErr))
			})
		})
	})

	Describe("GetOrganizations", func() {
		var (
			orgs     []Organization
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			orgs, warnings, err = actor.GetOrganizations()
		})

		When("there are multiple organizations", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{
						{
							Name: "some-org-1",
						},
						{
							Name: "some-org-2",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("returns the org and all warnings", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(orgs).To(HaveLen(2))
				Expect(orgs[0].Name).To(Equal("some-org-1"))
				Expect(orgs[1].Name).To(Equal("some-org-2"))

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				queriesArg := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(queriesArg).To(BeNil())
			})
		})

		When("there are no orgs", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns warnings and an empty list of orgs", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(orgs).To(HaveLen(0))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some get org error")
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{"warning-1", "warning-2"},
					expectedErr,
				)
			})

			It("returns warnings and the error", func() {
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})
