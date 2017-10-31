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

var _ = Describe("OrganizationQuota Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetOrganizationQuota", func() {
		Context("when the org quota exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotaReturns(
					ccv2.OrganizationQuota{
						GUID: "some-org-quota-guid",
						Name: "some-org-quota",
					},
					ccv2.Warnings{"warning-1"},
					nil,
				)
			})

			It("returns the org quota and warnings", func() {
				orgQuota, warnings, err := actor.GetOrganizationQuota("some-org-quota-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(orgQuota).To(Equal(OrganizationQuota{
					GUID: "some-org-quota-guid",
					Name: "some-org-quota",
				}))
				Expect(warnings).To(Equal(Warnings{"warning-1"}))

				Expect(fakeCloudControllerClient.GetOrganizationQuotaCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationQuotaArgsForCall(0)).To(Equal(
					"some-org-quota-guid"))
			})
		})

		Context("when the org quota does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotaReturns(ccv2.OrganizationQuota{}, nil, ccerror.ResourceNotFoundError{})
			})

			It("returns an OrganizationQuotaNotFoundError", func() {
				_, _, err := actor.GetOrganizationQuota("some-org-quota-guid")
				Expect(err).To(MatchError(actionerror.OrganizationQuotaNotFoundError{GUID: "some-org-quota-guid"}))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some org quota error")
				fakeCloudControllerClient.GetOrganizationQuotaReturns(ccv2.OrganizationQuota{}, nil, expectedErr)
			})

			It("returns the error", func() {
				_, _, err := actor.GetOrganizationQuota("some-org-quota-guid")
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})
})
