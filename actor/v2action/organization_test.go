package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Org Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
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

		Context("when the org exists", func() {
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

		Context("when the org does not exist", func() {
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

		Context("when client returns an error", func() {
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

		Context("when the org exists", func() {
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
				query := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(query).To(Equal(
					[]ccv2.QQuery{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-org"},
					}}))
			})
		})

		Context("when the org does not exist", func() {
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

		Context("when multiple orgs exist with the same name", func() {
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

		Context("when an error is encountered", func() {
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
					Status: ccv2.JobStatusFinished,
				}

				fakeCloudControllerClient.DeleteOrganizationReturns(
					job, ccv2.Warnings{"delete-org-warning"}, nil)

				fakeCloudControllerClient.PollJobReturns(ccv2.Warnings{"polling-warnings"}, nil)
			})

			It("returns warnings and deletes the org", func() {
				Expect(warnings).To(ConsistOf("get-org-warning", "delete-org-warning", "polling-warnings"))
				Expect(deleteOrgErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(query).To(Equal(
					[]ccv2.QQuery{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-org"},
					}}))

				Expect(fakeCloudControllerClient.DeleteOrganizationCallCount()).To(Equal(1))
				orgGuid := fakeCloudControllerClient.DeleteOrganizationArgsForCall(0)
				Expect(orgGuid).To(Equal("some-org-guid"))

				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				job := fakeCloudControllerClient.PollJobArgsForCall(0)
				Expect(job.GUID).To(Equal("some-job-guid"))
			})
		})

		Context("when getting the org returns an error", func() {
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

		Context("when the delete returns an error", func() {
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

				fakeCloudControllerClient.DeleteOrganizationReturns(
					ccv2.Job{},
					ccv2.Warnings{"delete-org-warning"},
					returnedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(deleteOrgErr).To(MatchError(returnedErr))
				Expect(warnings).To(ConsistOf("get-org-warning", "delete-org-warning"))
			})
		})

		Context("when the job polling has an error", func() {
			var expectedErr error
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv2.Organization{
					{GUID: "some-org-guid"},
				}, ccv2.Warnings{"get-org-warning"}, nil)

				fakeCloudControllerClient.DeleteOrganizationReturns(
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

		Context("when there are multiple organizations", func() {
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

		Context("when there are no orgs", func() {
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

		Context("when client returns an error", func() {
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
