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

var _ = Describe("SpaceQuota Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetSpaceQuota", func() {
		When("the space quota exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotaDefinitionReturns(
					ccv2.SpaceQuota{
						GUID: "some-space-quota-guid",
						Name: "some-space-quota",
					},
					ccv2.Warnings{"warning-1"},
					nil,
				)
			})

			It("returns the space quota and warnings", func() {
				spaceQuota, warnings, err := actor.GetSpaceQuota("some-space-quota-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(spaceQuota).To(Equal(SpaceQuota{
					GUID: "some-space-quota-guid",
					Name: "some-space-quota",
				}))
				Expect(warnings).To(ConsistOf("warning-1"))

				Expect(fakeCloudControllerClient.GetSpaceQuotaDefinitionCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpaceQuotaDefinitionArgsForCall(0)).To(Equal(
					"some-space-quota-guid"))
			})
		})

		When("the space quota does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotaDefinitionReturns(ccv2.SpaceQuota{}, nil, ccerror.ResourceNotFoundError{})
			})

			It("returns an SpaceQuotaNotFoundError", func() {
				_, _, err := actor.GetSpaceQuota("some-space-quota-guid")
				Expect(err).To(MatchError(actionerror.SpaceQuotaNotFoundError{GUID: "some-space-quota-guid"}))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some space quota error")
				fakeCloudControllerClient.GetSpaceQuotaDefinitionReturns(ccv2.SpaceQuota{}, ccv2.Warnings{"warning-1", "warning-2"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetSpaceQuota("some-space-quota-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetSpaceQuotaByName", func() {
		When("the orgGUID is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotasReturns(
					[]ccv2.SpaceQuota{},
					ccv2.Warnings{
						"warning-1",
						"warning-2",
					},
					actionerror.OrganizationNotFoundError{GUID: "some-org-guid"},
				)
			})

			It("returns the OrganizationNotFoundError", func() {
				spaceQuota, warnings, err := actor.GetSpaceQuotaByName("some-space-quota", "some-org-guid")
				Expect(spaceQuota).To(Equal(SpaceQuota{}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(err).To(MatchError(actionerror.OrganizationNotFoundError{GUID: "some-org-guid"}))
			})

		})

		When("the space quota is not found", func() {
			BeforeEach(func() {
				spaceQuota1 := ccv2.SpaceQuota{
					Name: "space-quota-1",
				}
				spaceQuota2 := ccv2.SpaceQuota{
					Name: "space-quota-2",
				}

				fakeCloudControllerClient.GetSpaceQuotasReturns(
					[]ccv2.SpaceQuota{
						spaceQuota1,
						spaceQuota2,
					},
					ccv2.Warnings{
						"warning-1",
						"warning-2",
					},
					nil,
				)
			})

			It("returns the SpaceQuotaNotFoundByNameError", func() {
				spaceQuota, warnings, err := actor.GetSpaceQuotaByName("some-space-quota", "some-org-guid")
				Expect(spaceQuota).To(Equal(SpaceQuota{}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(err).To(MatchError(actionerror.SpaceQuotaNotFoundByNameError{Name: "some-space-quota"}))
			})

		})

		When("the space quota is found", func() {
			BeforeEach(func() {
				spaceQuota1 := ccv2.SpaceQuota{
					Name: "space-quota-1",
				}
				spaceQuota2 := ccv2.SpaceQuota{
					Name: "space-quota-2",
				}

				fakeCloudControllerClient.GetSpaceQuotasReturns(
					[]ccv2.SpaceQuota{
						spaceQuota1,
						spaceQuota2,
					},
					ccv2.Warnings{
						"warning-1",
						"warning-2",
					},
					nil,
				)
			})

			It("returns the space quota", func() {
				spaceQuota, warnings, err := actor.GetSpaceQuotaByName("space-quota-2", "some-org-guid")
				Expect(spaceQuota).To(Equal(SpaceQuota{Name: "space-quota-2"}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("SetSpaceQuota", func() {
		When("the client call succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.SetSpaceQuotaReturns(
					ccv2.Warnings{
						"warning-1",
						"warning-2",
					},
					nil,
				)
			})

			It("sets the space quota and returns the warnings", func() {
				warnings, err := actor.SetSpaceQuota("some-space-guid", "some-quota-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("the client call fails", func() {

		})
	})
})
