package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space Quota Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		trueValue                 = true
		falseValue                = true
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _, _ = NewTestActor()
	})

	Describe("ApplySpaceQuotaByName", func() {
		var (
			warnings   Warnings
			executeErr error
			quotaName  = "space-quota-name"
			quotaGUID  = "space-quota-guid"
			spaceGUID  = "space-guid"
			orgGUID    = "org-guid"
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.ApplySpaceQuotaByName(quotaName, spaceGUID, orgGUID)
		})

		When("the space quota could not be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotasReturns(
					[]resources.SpaceQuota{},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetSpaceQuotasCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.ApplySpaceQuotaCallCount()).To(Equal(0))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(executeErr).To(MatchError(actionerror.SpaceQuotaNotFoundForNameError{Name: quotaName}))
			})
		})

		When("applying the quota returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotasReturns(
					[]resources.SpaceQuota{
						{Quota: resources.Quota{GUID: "some-quota-guid"}},
					},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
				fakeCloudControllerClient.ApplySpaceQuotaReturns(
					resources.RelationshipList{},
					ccv3.Warnings{"apply-quota-warning"},
					errors.New("apply-quota-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetSpaceQuotasCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.ApplySpaceQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning", "apply-quota-warning"))
				Expect(executeErr).To(MatchError("apply-quota-error"))
			})
		})

		When("the quota is successfully applied to the space", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotasReturns(
					[]resources.SpaceQuota{
						{Quota: resources.Quota{GUID: quotaGUID}},
					},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
				fakeCloudControllerClient.ApplySpaceQuotaReturns(
					resources.RelationshipList{
						GUIDs: []string{orgGUID},
					},
					ccv3.Warnings{"apply-quota-warning"},
					nil,
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetSpaceQuotasCallCount()).To(Equal(1))
				passedQuotaQuery := fakeCloudControllerClient.GetSpaceQuotasArgsForCall(0)
				Expect(passedQuotaQuery).To(Equal(
					[]ccv3.Query{
						{
							Key:    "organization_guids",
							Values: []string{orgGUID},
						},
						{
							Key:    "names",
							Values: []string{quotaName},
						},
					},
				))

				Expect(fakeCloudControllerClient.ApplySpaceQuotaCallCount()).To(Equal(1))
				passedQuotaGUID, passedSpaceGUID := fakeCloudControllerClient.ApplySpaceQuotaArgsForCall(0)
				Expect(passedQuotaGUID).To(Equal(quotaGUID))
				Expect(passedSpaceGUID).To(Equal(spaceGUID))

				Expect(warnings).To(ConsistOf("some-quota-warning", "apply-quota-warning"))
				Expect(executeErr).To(BeNil())
			})
		})
	})

	Describe("CreateSpaceQuota", func() {
		var (
			spaceQuotaName   string
			organizationGuid string
			warnings         Warnings
			executeErr       error
			limits           QuotaLimits
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.CreateSpaceQuota(spaceQuotaName, organizationGuid, limits)
		})

		When("creating a space quota with all values set", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateSpaceQuotaReturns(resources.SpaceQuota{}, ccv3.Warnings{"some-quota-warning"}, nil)
				limits = QuotaLimits{
					TotalMemoryInMB:       &types.NullInt{IsSet: true, Value: 2},
					PerProcessMemoryInMB:  &types.NullInt{IsSet: true, Value: 3},
					TotalInstances:        &types.NullInt{IsSet: true, Value: 4},
					PaidServicesAllowed:   &trueValue,
					TotalServiceInstances: &types.NullInt{IsSet: true, Value: 6},
					TotalRoutes:           &types.NullInt{IsSet: true, Value: 8},
					TotalReservedPorts:    &types.NullInt{IsSet: true, Value: 9},
					TotalLogVolume:        &types.NullInt{IsSet: true, Value: 10},
				}
			})

			It("makes the space quota", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.CreateSpaceQuotaCallCount()).To(Equal(1))
				givenSpaceQuota := fakeCloudControllerClient.CreateSpaceQuotaArgsForCall(0)

				Expect(givenSpaceQuota).To(Equal(resources.SpaceQuota{
					Quota: resources.Quota{
						Name: spaceQuotaName,
						Apps: resources.AppLimit{
							TotalMemory:       &types.NullInt{IsSet: true, Value: 2},
							InstanceMemory:    &types.NullInt{IsSet: true, Value: 3},
							TotalAppInstances: &types.NullInt{IsSet: true, Value: 4},
							TotalLogVolume:    &types.NullInt{IsSet: true, Value: 10},
						},
						Services: resources.ServiceLimit{
							TotalServiceInstances: &types.NullInt{IsSet: true, Value: 6},
							PaidServicePlans:      &trueValue,
						},
						Routes: resources.RouteLimit{
							TotalRoutes:        &types.NullInt{IsSet: true, Value: 8},
							TotalReservedPorts: &types.NullInt{IsSet: true, Value: 9},
						},
					},
					OrgGUID:    organizationGuid,
					SpaceGUIDs: nil,
				}))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
			})
		})

		When("creating a space quota with empty limits", func() {
			var (
				ccv3Quota resources.SpaceQuota
			)

			BeforeEach(func() {
				spaceQuotaName = "quota-name"
				limits = QuotaLimits{}

				ccv3Quota = resources.SpaceQuota{
					Quota: resources.Quota{
						Name: spaceQuotaName,
						Apps: resources.AppLimit{
							TotalMemory:       &types.NullInt{Value: 0, IsSet: true},
							InstanceMemory:    nil,
							TotalAppInstances: nil,
							TotalLogVolume:    nil,
						},
						Services: resources.ServiceLimit{
							TotalServiceInstances: &types.NullInt{Value: 0, IsSet: true},
							PaidServicePlans:      nil,
						},
						Routes: resources.RouteLimit{
							TotalRoutes:        &types.NullInt{Value: 0, IsSet: true},
							TotalReservedPorts: &types.NullInt{Value: 0, IsSet: true},
						},
					},
				}
				fakeCloudControllerClient.CreateSpaceQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("call the create endpoint with the respective values and returns warnings", func() {
				Expect(fakeCloudControllerClient.CreateSpaceQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))

				passedQuota := fakeCloudControllerClient.CreateSpaceQuotaArgsForCall(0)
				Expect(passedQuota).To(Equal(ccv3Quota))
			})
		})

		When("creating a quota with all values set to unlimited", func() {
			var (
				ccv3Quota resources.SpaceQuota
			)

			BeforeEach(func() {
				spaceQuotaName = "quota-name"
				limits = QuotaLimits{
					TotalMemoryInMB:       &types.NullInt{Value: -1, IsSet: true},
					PerProcessMemoryInMB:  &types.NullInt{Value: -1, IsSet: true},
					TotalInstances:        &types.NullInt{Value: -1, IsSet: true},
					PaidServicesAllowed:   &trueValue,
					TotalServiceInstances: &types.NullInt{Value: -1, IsSet: true},
					TotalRoutes:           &types.NullInt{Value: -1, IsSet: true},
					TotalReservedPorts:    &types.NullInt{Value: -1, IsSet: true},
					TotalLogVolume:        &types.NullInt{Value: -1, IsSet: true},
				}
				ccv3Quota = resources.SpaceQuota{
					Quota: resources.Quota{
						Name: spaceQuotaName,
						Apps: resources.AppLimit{
							TotalMemory:       &types.NullInt{Value: 0, IsSet: false},
							InstanceMemory:    &types.NullInt{Value: 0, IsSet: false},
							TotalAppInstances: &types.NullInt{Value: 0, IsSet: false},
							TotalLogVolume:    &types.NullInt{Value: 0, IsSet: false},
						},
						Services: resources.ServiceLimit{
							TotalServiceInstances: &types.NullInt{Value: 0, IsSet: false},
							PaidServicePlans:      &trueValue,
						},
						Routes: resources.RouteLimit{
							TotalRoutes:        &types.NullInt{Value: 0, IsSet: false},
							TotalReservedPorts: &types.NullInt{Value: 0, IsSet: false},
						},
					},
				}
				fakeCloudControllerClient.CreateSpaceQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("call the create endpoint with the respective values and returns warnings", func() {
				Expect(fakeCloudControllerClient.CreateSpaceQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))

				passedQuota := fakeCloudControllerClient.CreateSpaceQuotaArgsForCall(0)
				Expect(passedQuota).To(Equal(ccv3Quota))
			})
		})

		When("creating a quota returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateSpaceQuotaReturns(
					resources.SpaceQuota{},
					ccv3.Warnings{"some-quota-warning"},
					errors.New("create-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(fakeCloudControllerClient.CreateSpaceQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(executeErr).To(MatchError("create-error"))
			})
		})
	})

	Describe("DeleteSpaceQuotaByName", func() {
		var (
			quotaName  string
			orgGUID    string
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			quotaName = "quota-name"
			orgGUID = "some-org-guid"

			fakeCloudControllerClient.GetSpaceQuotasReturns(
				[]resources.SpaceQuota{{Quota: resources.Quota{GUID: "some-quota-guid"}}},
				ccv3.Warnings{"get-quota-warning"},
				nil,
			)

			fakeCloudControllerClient.DeleteSpaceQuotaReturns(
				ccv3.JobURL("some-job-url"),
				ccv3.Warnings{"delete-quota-warning"},
				nil,
			)

			fakeCloudControllerClient.PollJobReturns(
				ccv3.Warnings{"poll-job-warning"},
				nil,
			)
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.DeleteSpaceQuotaByName(quotaName, orgGUID)
		})

		When("no errors occur", func() {
			It("retrieves the space quota by name, makes the API call, and polls the deletion job until completion", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-quota-warning", "delete-quota-warning", "poll-job-warning"))

				Expect(fakeCloudControllerClient.GetSpaceQuotasCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetSpaceQuotasArgsForCall(0)
				Expect(query).To(ConsistOf(
					ccv3.Query{
						Key:    ccv3.OrganizationGUIDFilter,
						Values: []string{orgGUID},
					},
					ccv3.Query{
						Key:    ccv3.NameFilter,
						Values: []string{quotaName},
					},
				))

				Expect(fakeCloudControllerClient.DeleteSpaceQuotaCallCount()).To(Equal(1))
				quotaGUID := fakeCloudControllerClient.DeleteSpaceQuotaArgsForCall(0)
				Expect(quotaGUID).To(Equal("some-quota-guid"))

				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				inputJobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
				Expect(inputJobURL).To(Equal(ccv3.JobURL("some-job-url")))
			})
		})

		When("there is an error getting the space quota", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotasReturns(
					[]resources.SpaceQuota{},
					ccv3.Warnings{"get-quota-warning"},
					errors.New("get-quota-error"),
				)
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("get-quota-error"))
				Expect(warnings).To(ConsistOf("get-quota-warning"))
			})
		})

		When("there is an error deleting the space quota", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteSpaceQuotaReturns(
					"",
					ccv3.Warnings{"delete-quota-warning"},
					errors.New("delete-quota-error"),
				)
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("delete-quota-error"))
				Expect(warnings).To(ConsistOf("get-quota-warning", "delete-quota-warning"))
			})
		})

		When("there is an error polling the job", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PollJobReturns(
					ccv3.Warnings{"poll-job-warning"},
					errors.New("poll-job-error"),
				)
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("poll-job-error"))
				Expect(warnings).To(ConsistOf("get-quota-warning", "delete-quota-warning", "poll-job-warning"))
			})
		})
	})

	Describe("GetSpaceQuotaByName", func() {
		var (
			quotaName  string
			orgGUID    string
			quota      resources.SpaceQuota
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			quotaName = "quota-name"
			orgGUID = "some-org-guid"
		})

		JustBeforeEach(func() {
			quota, warnings, executeErr = actor.GetSpaceQuotaByName(quotaName, orgGUID)
		})

		When("when the API layer call returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotasReturns(
					[]resources.SpaceQuota{},
					ccv3.Warnings{"some-quota-warning"},
					errors.New("list-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetSpaceQuotasCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(executeErr).To(MatchError("list-error"))
				Expect(quota).To(Equal(resources.SpaceQuota{}))
			})
		})

		When("when the space quota could not be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotasReturns(
					[]resources.SpaceQuota{},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetSpaceQuotasCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(executeErr).To(MatchError(actionerror.SpaceQuotaNotFoundForNameError{Name: quotaName}))
				Expect(quota).To(Equal(resources.SpaceQuota{}))
			})
		})

		When("getting a single quota by name", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotasReturns(
					[]resources.SpaceQuota{
						{
							Quota: resources.Quota{
								GUID: "quota-guid",
								Name: quotaName,
							},
						},
					},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("queries the API and returns the matching space quota", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetSpaceQuotasCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetSpaceQuotasArgsForCall(0)
				Expect(query).To(ConsistOf(
					ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{quotaName}},
				))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(quota).To(Equal(resources.SpaceQuota{
					Quota: resources.Quota{
						GUID: "quota-guid",
						Name: quotaName,
					},
				}))
			})
		})
	})

	Describe("GetSpaceQuotasByOrgGUID", func() {
		var (
			orgGUID    string
			quotas     []resources.SpaceQuota
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			orgGUID = "org-guid"
		})

		JustBeforeEach(func() {
			quotas, warnings, executeErr = actor.GetSpaceQuotasByOrgGUID(orgGUID)
		})

		When("when the API layer call returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotasReturns(
					[]resources.SpaceQuota{},
					ccv3.Warnings{"some-quota-warning"},
					errors.New("list-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetSpaceQuotasCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(executeErr).To(MatchError("list-error"))
				Expect(quotas).To(Equal([]resources.SpaceQuota{}))
			})
		})

		When("getting all space quotas associated with the same organization", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceQuotasReturns(
					[]resources.SpaceQuota{
						{
							Quota: resources.Quota{
								GUID: "quota-guid",
								Name: "quota-beluga",
							},
							OrgGUID: orgGUID,
						},
						{
							Quota: resources.Quota{
								GUID: "quota-2-guid",
								Name: "quota-manatee",
							},
							OrgGUID: orgGUID,
						},
					},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("queries the API and returns the matching space quotas", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetSpaceQuotasCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetSpaceQuotasArgsForCall(0)
				Expect(query).To(ConsistOf(
					ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
				))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(quotas).To(ConsistOf(
					resources.SpaceQuota{
						Quota: resources.Quota{
							GUID: "quota-guid",
							Name: "quota-beluga",
						},
						OrgGUID: orgGUID,
					},
					resources.SpaceQuota{
						Quota: resources.Quota{
							GUID: "quota-2-guid",
							Name: "quota-manatee",
						},
						OrgGUID: orgGUID,
					},
				))
			})
		})
	})

	Describe("UpdateSpaceQuota", func() {
		var (
			oldQuotaName string
			orgGUID      string
			newQuotaName string
			quotaLimits  QuotaLimits
			warnings     Warnings
			executeErr   error
		)

		BeforeEach(func() {
			oldQuotaName = "old-quota-name"
			orgGUID = "some-org-guid"
			newQuotaName = "new-quota-name"

			quotaLimits = QuotaLimits{
				TotalMemoryInMB:       &types.NullInt{Value: 2048, IsSet: true},
				PerProcessMemoryInMB:  &types.NullInt{Value: 1024, IsSet: true},
				TotalInstances:        &types.NullInt{Value: 0, IsSet: false},
				TotalServiceInstances: &types.NullInt{Value: 0, IsSet: true},
				PaidServicesAllowed:   &trueValue,
				TotalRoutes:           &types.NullInt{Value: 6, IsSet: true},
				TotalReservedPorts:    &types.NullInt{Value: 5, IsSet: true},
				TotalLogVolume:        &types.NullInt{Value: 512, IsSet: true},
			}

			fakeCloudControllerClient.GetSpaceQuotasReturns(
				[]resources.SpaceQuota{{Quota: resources.Quota{Name: oldQuotaName}}},
				ccv3.Warnings{"get-quotas-warning"},
				nil,
			)
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateSpaceQuota(oldQuotaName, orgGUID, newQuotaName, quotaLimits)
		})

		When("the update-quota endpoint returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateSpaceQuotaReturns(
					resources.SpaceQuota{},
					ccv3.Warnings{"update-quota-warning"},
					errors.New("update-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(fakeCloudControllerClient.UpdateSpaceQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("get-quotas-warning", "update-quota-warning"))
				Expect(executeErr).To(MatchError("update-error"))
			})
		})

		When("no quota limits are being updated", func() {
			var (
				ccv3Quota resources.SpaceQuota
			)

			BeforeEach(func() {
				quotaLimits = QuotaLimits{}

				ccv3Quota = resources.SpaceQuota{
					Quota: resources.Quota{
						Name: oldQuotaName,
						Apps: resources.AppLimit{
							TotalMemory:       nil,
							InstanceMemory:    nil,
							TotalAppInstances: nil,
							TotalLogVolume:    nil,
						},
						Services: resources.ServiceLimit{
							TotalServiceInstances: nil,
							PaidServicePlans:      nil,
						},
						Routes: resources.RouteLimit{
							TotalRoutes:        nil,
							TotalReservedPorts: nil,
						},
					},
				}

				fakeCloudControllerClient.UpdateSpaceQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"update-quota-warning"},
					nil,
				)
			})

			It("calls the update endpoint with the respective values and returns warnings", func() {
				Expect(fakeCloudControllerClient.UpdateSpaceQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("get-quotas-warning", "update-quota-warning"))

				passedQuota := fakeCloudControllerClient.UpdateSpaceQuotaArgsForCall(0)

				updatedQuota := ccv3Quota
				updatedQuota.Name = newQuotaName

				Expect(passedQuota).To(Equal(updatedQuota))
			})
		})

		When("the update space quota has all values set to unlimited", func() {
			var (
				ccv3Quota resources.SpaceQuota
			)

			BeforeEach(func() {
				quotaLimits = QuotaLimits{
					TotalMemoryInMB:       &types.NullInt{Value: -1, IsSet: true},
					PerProcessMemoryInMB:  &types.NullInt{Value: -1, IsSet: true},
					TotalInstances:        &types.NullInt{Value: -1, IsSet: true},
					PaidServicesAllowed:   &falseValue,
					TotalServiceInstances: &types.NullInt{Value: -1, IsSet: true},
					TotalRoutes:           &types.NullInt{Value: -1, IsSet: true},
					TotalReservedPorts:    &types.NullInt{Value: -1, IsSet: true},
					TotalLogVolume:        &types.NullInt{Value: -1, IsSet: true},
				}

				ccv3Quota = resources.SpaceQuota{
					Quota: resources.Quota{
						Name: oldQuotaName,
						Apps: resources.AppLimit{
							TotalMemory:       &types.NullInt{Value: 0, IsSet: false},
							InstanceMemory:    &types.NullInt{Value: 0, IsSet: false},
							TotalAppInstances: &types.NullInt{Value: 0, IsSet: false},
							TotalLogVolume:    &types.NullInt{Value: 0, IsSet: false},
						},
						Services: resources.ServiceLimit{
							TotalServiceInstances: &types.NullInt{Value: 0, IsSet: false},
							PaidServicePlans:      &falseValue,
						},
						Routes: resources.RouteLimit{
							TotalRoutes:        &types.NullInt{Value: 0, IsSet: false},
							TotalReservedPorts: &types.NullInt{Value: 0, IsSet: false},
						},
					},
				}

				fakeCloudControllerClient.UpdateSpaceQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"update-quota-warning"},
					nil,
				)
			})

			It("calls the update endpoint with the respective values and returns warnings", func() {
				Expect(fakeCloudControllerClient.UpdateSpaceQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("get-quotas-warning", "update-quota-warning"))

				passedQuota := fakeCloudControllerClient.UpdateSpaceQuotaArgsForCall(0)

				updatedQuota := ccv3Quota
				updatedQuota.Name = newQuotaName

				Expect(passedQuota).To(Equal(updatedQuota))
			})
		})

		When("The update space quota endpoint succeeds", func() {
			var (
				ccv3Quota resources.SpaceQuota
			)

			BeforeEach(func() {
				ccv3Quota = resources.SpaceQuota{
					Quota: resources.Quota{
						Name: oldQuotaName,
						Apps: resources.AppLimit{
							TotalMemory:       &types.NullInt{Value: 2048, IsSet: true},
							InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
							TotalAppInstances: &types.NullInt{Value: 0, IsSet: false},
							TotalLogVolume:    &types.NullInt{Value: 512, IsSet: true},
						},
						Services: resources.ServiceLimit{
							TotalServiceInstances: &types.NullInt{Value: 0, IsSet: true},
							PaidServicePlans:      &trueValue,
						},
						Routes: resources.RouteLimit{
							TotalRoutes:        &types.NullInt{Value: 6, IsSet: true},
							TotalReservedPorts: &types.NullInt{Value: 5, IsSet: true},
						},
					},
				}

				fakeCloudControllerClient.UpdateSpaceQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"update-quota-warning"},
					nil,
				)
			})

			It("calls the update endpoint with the respective values and returns warnings", func() {
				Expect(fakeCloudControllerClient.UpdateSpaceQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("get-quotas-warning", "update-quota-warning"))

				passedQuota := fakeCloudControllerClient.UpdateSpaceQuotaArgsForCall(0)

				updatedQuota := ccv3Quota
				updatedQuota.Name = newQuotaName

				Expect(passedQuota).To(Equal(updatedQuota))
			})
		})

		When("the space quota name is not being updated", func() {
			var (
				ccv3Quota resources.SpaceQuota
			)

			BeforeEach(func() {
				newQuotaName = ""

				ccv3Quota = resources.SpaceQuota{
					Quota: resources.Quota{
						Name: oldQuotaName,
						Apps: resources.AppLimit{
							TotalMemory:       &types.NullInt{Value: 2048, IsSet: true},
							InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
							TotalAppInstances: &types.NullInt{Value: 0, IsSet: false},
							TotalLogVolume:    &types.NullInt{Value: 512, IsSet: true},
						},
						Services: resources.ServiceLimit{
							TotalServiceInstances: &types.NullInt{Value: 0, IsSet: true},
							PaidServicePlans:      &trueValue,
						},
						Routes: resources.RouteLimit{
							TotalRoutes:        &types.NullInt{Value: 6, IsSet: true},
							TotalReservedPorts: &types.NullInt{Value: 5, IsSet: true},
						},
					},
				}

				fakeCloudControllerClient.UpdateSpaceQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"update-quota-warning"},
					nil,
				)
			})
			It("uses the current space quota name in the API request", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				inputQuota := fakeCloudControllerClient.UpdateSpaceQuotaArgsForCall(0)
				Expect(inputQuota.Name).To(Equal("old-quota-name"))
			})
		})
	})

	Describe("UnsetSpaceQuota", func() {
		var (
			spaceQuotaName string
			orgGUID        string
			spaceName      string
			warnings       Warnings
			executeErr     error
		)

		BeforeEach(func() {
			spaceQuotaName = "some-quota-name"
			orgGUID = "some-org-guid"
			spaceName = "some-space-name"

			fakeCloudControllerClient.GetSpaceQuotasReturns(
				[]resources.SpaceQuota{{Quota: resources.Quota{Name: spaceQuotaName}}},
				ccv3.Warnings{"get-quotas-warning"},
				nil,
			)

			fakeCloudControllerClient.GetSpacesReturns(
				[]resources.Space{{Name: spaceName}},
				ccv3.IncludedResources{},
				ccv3.Warnings{"get-spaces-warning"},
				nil,
			)
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.UnsetSpaceQuota(spaceQuotaName, spaceName, orgGUID)
		})

		When("getting the space fails", func() {
			When("no space with that name exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]resources.Space{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get-spaces-warning"},
						nil,
					)
				})

				It("returns the error and prints warnings", func() {
					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))

					Expect(warnings).To(ContainElement("get-spaces-warning"))
					Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: spaceName}))
				})
			})

			When("getting the space returns an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]resources.Space{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get-spaces-warning"},
						errors.New("some-get-spaces-error"),
					)
				})

				It("returns the error and prints warnings", func() {
					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))

					Expect(warnings).To(ContainElement("get-spaces-warning"))
					Expect(executeErr).To(MatchError("some-get-spaces-error"))
				})
			})
		})

		When("getting the space quota fails", func() {
			When("no space with that name exists", func() {

				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceQuotasReturns(
						[]resources.SpaceQuota{},
						ccv3.Warnings{"get-quota-warning"},
						nil,
					)
				})

				It("returns the error and prints warnings", func() {
					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))

					Expect(warnings).To(ContainElement("get-quota-warning"))
					Expect(executeErr).To(MatchError(actionerror.SpaceQuotaNotFoundForNameError{Name: spaceQuotaName}))
				})
			})

			When("getting the space returns an error", func() {

				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceQuotasReturns(
						[]resources.SpaceQuota{},
						ccv3.Warnings{"get-quota-warning"},
						errors.New("some-get-quotas-error"),
					)
				})

				It("returns the error and prints warnings", func() {
					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))

					Expect(warnings).To(ContainElement("get-quota-warning"))
					Expect(executeErr).To(MatchError("some-get-quotas-error"))
				})
			})
		})

		When("unsetting the space quota fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UnsetSpaceQuotaReturns(
					ccv3.Warnings{"unset-quota-warning"},
					errors.New("some-unset-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(warnings).To(ConsistOf("unset-quota-warning", "get-spaces-warning", "get-quotas-warning"))
				Expect(executeErr).To(MatchError("some-unset-error"))
			})
		})

	})

})
