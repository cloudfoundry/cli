package v7action_test

import (
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space Quota Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _ = NewTestActor()

	})

	Describe("CreateSpaceQuota", func() {
		var (
			spaceQuotaName   string
			organizationGuid string
			warnings         Warnings
			executeErr       error
			limits           SpaceQuotaLimits
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.CreateSpaceQuota(spaceQuotaName, organizationGuid, limits)
		})

		When("creating a space quota", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateSpaceQuotaReturns(ccv3.SpaceQuota{}, ccv3.Warnings{"some-quota-warning"}, nil)
				limits = SpaceQuotaLimits{
					TotalMemoryInMB:       types.NullInt{IsSet: true, Value: 2},
					PerProcessMemoryInMB:  types.NullInt{IsSet: true, Value: 3},
					TotalInstances:        types.NullInt{IsSet: true, Value: 4},
					PerAppTasks:           types.NullInt{IsSet: true, Value: 5},
					PaidServicesAllowed:   true,
					TotalServiceInstances: types.NullInt{IsSet: true, Value: 6},
					TotalServiceKeys:      types.NullInt{IsSet: true, Value: 7},
					TotalRoutes:           types.NullInt{IsSet: true, Value: 8},
					TotalReservedPorts:    types.NullInt{IsSet: true, Value: 9},
				}
			})

			It("makes the space quota", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.CreateSpaceQuotaCallCount()).To(Equal(1))
				givenSpaceQuota := fakeCloudControllerClient.CreateSpaceQuotaArgsForCall(0)

				Expect(givenSpaceQuota).To(Equal(ccv3.SpaceQuota{
					Name:    spaceQuotaName,
					OrgGUID: organizationGuid,
					Apps: ccv3.AppLimit{
						TotalMemory:       types.NullInt{IsSet: true, Value: 2},
						InstanceMemory:    types.NullInt{IsSet: true, Value: 3},
						TotalAppInstances: types.NullInt{IsSet: true, Value: 4},
					},
					Services: ccv3.ServiceLimit{
						TotalServiceInstances: types.NullInt{IsSet: true, Value: 6},
						PaidServicePlans:      true,
					},
					Routes: ccv3.RouteLimit{
						TotalRoutes:     types.NullInt{IsSet: true, Value: 8},
						TotalRoutePorts: types.NullInt{IsSet: true, Value: 9},
					},
					SpaceGUIDs: nil,
				}))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
			})
		})
	})
})
