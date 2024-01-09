package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Organization Quota Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		trueValue                 bool = true
		falseValue                bool = false
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _, _ = NewTestActor()
	})

	Describe("ApplyOrganizationQuotaByName", func() {
		var (
			warnings   Warnings
			executeErr error
			quotaName  = "org-quota-name"
			orgGUID    = "org-guid"
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.ApplyOrganizationQuotaByName(quotaName, orgGUID)
		})

		When("when the org quota could not be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]resources.OrganizationQuota{},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(executeErr).To(MatchError(actionerror.OrganizationQuotaNotFoundForNameError{Name: quotaName}))
			})
		})

		When("when applying the quota returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]resources.OrganizationQuota{
						{Quota: resources.Quota{GUID: "some-quota-guid"}},
					},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
				fakeCloudControllerClient.ApplyOrganizationQuotaReturns(
					resources.RelationshipList{},
					ccv3.Warnings{"apply-quota-warning"},
					errors.New("apply-quota-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.ApplyOrganizationQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning", "apply-quota-warning"))
				Expect(executeErr).To(MatchError("apply-quota-error"))
			})
		})

		When("Quota is successfully applied to the org", func() {
			var quotaGUID = "some-quota-guid"
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]resources.OrganizationQuota{
						{Quota: resources.Quota{GUID: quotaGUID}},
					},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
				fakeCloudControllerClient.ApplyOrganizationQuotaReturns(
					resources.RelationshipList{
						GUIDs: []string{orgGUID},
					},
					ccv3.Warnings{"apply-quota-warning"},
					nil,
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))
				passedQuotaQuery := fakeCloudControllerClient.GetOrganizationQuotasArgsForCall(0)
				Expect(passedQuotaQuery).To(Equal([]ccv3.Query{
					{Key: ccv3.NameFilter, Values: []string{quotaName}},
					{Key: ccv3.PerPage, Values: []string{"1"}},
					{Key: ccv3.Page, Values: []string{"1"}},
				}))
				Expect(fakeCloudControllerClient.ApplyOrganizationQuotaCallCount()).To(Equal(1))
				passedQuotaGUID, passedOrgGUID := fakeCloudControllerClient.ApplyOrganizationQuotaArgsForCall(0)
				Expect(passedQuotaGUID).To(Equal(quotaGUID))
				Expect(passedOrgGUID).To(Equal(orgGUID))

				Expect(warnings).To(ConsistOf("some-quota-warning", "apply-quota-warning"))
				Expect(executeErr).To(BeNil())
			})
		})
	})

	Describe("DeleteOrganizationQuota", func() {
		var (
			quotaName  string
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			quotaName = "quota-name"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.DeleteOrganizationQuota(quotaName)
		})

		When("all API calls succeed", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]resources.OrganizationQuota{{Quota: resources.Quota{Name: quotaName, GUID: "quota-guid"}}},
					ccv3.Warnings{"get-quotas-warning"},
					nil,
				)

				fakeCloudControllerClient.DeleteOrganizationQuotaReturns(
					"some-job-url",
					ccv3.Warnings{"delete-quota-warning"},
					nil,
				)

				fakeCloudControllerClient.PollJobReturns(
					ccv3.Warnings{"poll-job-warning"},
					nil,
				)
			})

			It("returns warnings but no error", func() {
				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetOrganizationQuotasArgsForCall(0)
				Expect(query).To(Equal([]ccv3.Query{
					{Key: ccv3.NameFilter, Values: []string{quotaName}},
					{Key: ccv3.PerPage, Values: []string{"1"}},
					{Key: ccv3.Page, Values: []string{"1"}},
				}))

				Expect(fakeCloudControllerClient.DeleteOrganizationQuotaCallCount()).To(Equal(1))
				givenQuotaGUID := fakeCloudControllerClient.DeleteOrganizationQuotaArgsForCall(0)
				Expect(givenQuotaGUID).To(Equal("quota-guid"))

				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				givenJobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
				Expect(givenJobURL).To(Equal(ccv3.JobURL("some-job-url")))

				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-quotas-warning", "delete-quota-warning", "poll-job-warning"))
			})
		})

		When("getting the quota by name fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]resources.OrganizationQuota{{Quota: resources.Quota{Name: quotaName, GUID: "quota-guid"}}},
					ccv3.Warnings{"get-quotas-warning"},
					nil,
				)

				fakeCloudControllerClient.DeleteOrganizationQuotaReturns(
					"some-job-url",
					ccv3.Warnings{"delete-quota-warning"},
					errors.New("delete-quota-error"),
				)
			})

			It("returns error and warnings", func() {
				Expect(executeErr).To(MatchError("delete-quota-error"))
				Expect(warnings).To(ConsistOf("get-quotas-warning", "delete-quota-warning"))
			})
		})

		When("issuing the delete-quota request fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]resources.OrganizationQuota{{Quota: resources.Quota{Name: quotaName, GUID: "quota-guid"}}},
					ccv3.Warnings{"get-quotas-warning"},
					nil,
				)

				fakeCloudControllerClient.DeleteOrganizationQuotaReturns(
					"some-job-url",
					ccv3.Warnings{"delete-quota-warning"},
					nil,
				)

				fakeCloudControllerClient.PollJobReturns(
					ccv3.Warnings{"poll-job-warning"},
					errors.New("poll-job-error"),
				)
			})

			It("returns error and warnings", func() {
				Expect(executeErr).To(MatchError("poll-job-error"))
				Expect(warnings).To(ConsistOf("get-quotas-warning", "delete-quota-warning", "poll-job-warning"))
			})
		})

		When("the delete job fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]resources.OrganizationQuota{{Quota: resources.Quota{Name: quotaName, GUID: "quota-guid"}}},
					ccv3.Warnings{"get-quotas-warning"},
					errors.New("get-quotas-error"),
				)
			})

			It("returns error and warnings", func() {
				Expect(executeErr).To(MatchError("get-quotas-error"))
				Expect(warnings).To(ConsistOf("get-quotas-warning"))
			})
		})
	})

	Describe("GetOrganizationQuotas", func() {
		var (
			quotas     []resources.OrganizationQuota
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			quotas, warnings, executeErr = actor.GetOrganizationQuotas()
		})

		When("getting organization quotas", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]resources.OrganizationQuota{
						{
							Quota: resources.Quota{
								GUID: "quota-guid",
								Name: "kiwi",
							},
						},
						{
							Quota: resources.Quota{
								GUID: "quota-2-guid",
								Name: "strawberry",
							},
						},
					},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("queries the API and returns organization quotas", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))

				Expect(quotas).To(ConsistOf(
					resources.OrganizationQuota{
						Quota: resources.Quota{
							GUID: "quota-guid",
							Name: "kiwi",
						},
					},
					resources.OrganizationQuota{
						Quota: resources.Quota{
							GUID: "quota-2-guid",
							Name: "strawberry",
						},
					},
				))
				Expect(warnings).To(ConsistOf("some-quota-warning"))
			})
		})
	})

	Describe("GetOrganizationQuotaByName", func() {
		var (
			quotaName  string
			quota      resources.OrganizationQuota
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			quotaName = "quota-name"
		})

		JustBeforeEach(func() {
			quota, warnings, executeErr = actor.GetOrganizationQuotaByName(quotaName)
		})

		When("when the API layer call returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]resources.OrganizationQuota{},
					ccv3.Warnings{"some-quota-warning"},
					errors.New("list-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(executeErr).To(MatchError("list-error"))
				Expect(quota).To(Equal(resources.OrganizationQuota{}))
			})
		})

		When("when the org quota could not be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]resources.OrganizationQuota{},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(executeErr).To(MatchError(actionerror.OrganizationQuotaNotFoundForNameError{Name: quotaName}))
				Expect(quota).To(Equal(resources.OrganizationQuota{}))
			})
		})

		When("getting a single quota by name", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]resources.OrganizationQuota{
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

			It("queries the API and returns the matching organization quota", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetOrganizationQuotasArgsForCall(0)
				Expect(query).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{quotaName}},
					ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
					ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
				))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(quota).To(Equal(resources.OrganizationQuota{
					Quota: resources.Quota{
						GUID: "quota-guid",
						Name: quotaName,
					},
				}))
			})
		})
	})

	Describe("CreateOrganizationQuota", func() {
		var (
			quotaName   string
			quotaLimits QuotaLimits
			warnings    Warnings
			executeErr  error
		)

		BeforeEach(func() {
			quotaName = "quota-name"
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
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.CreateOrganizationQuota(quotaName, quotaLimits)
		})

		When("The create org v7Quota endpoint returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateOrganizationQuotaReturns(
					resources.OrganizationQuota{},
					ccv3.Warnings{"some-quota-warning"},
					errors.New("create-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(fakeCloudControllerClient.CreateOrganizationQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(executeErr).To(MatchError("create-error"))
			})
		})

		When("The create org quota has an empty org quota request", func() {
			var (
				ccv3Quota resources.OrganizationQuota
			)
			BeforeEach(func() {
				quotaName = "quota-name"
				quotaLimits = QuotaLimits{}

				ccv3Quota = resources.OrganizationQuota{
					Quota: resources.Quota{
						Name: quotaName,
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
				fakeCloudControllerClient.CreateOrganizationQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("call the create endpoint with the respective values and returns warnings", func() {
				Expect(fakeCloudControllerClient.CreateOrganizationQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))

				passedQuota := fakeCloudControllerClient.CreateOrganizationQuotaArgsForCall(0)
				Expect(passedQuota).To(Equal(ccv3Quota))
			})
		})

		When("The create org quota has all values set to unlimited", func() {
			var (
				ccv3Quota resources.OrganizationQuota
			)
			BeforeEach(func() {
				quotaName = "quota-name"
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
				ccv3Quota = resources.OrganizationQuota{
					Quota: resources.Quota{
						Name: quotaName,
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
				fakeCloudControllerClient.CreateOrganizationQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("calls the create endpoint with the respective values and returns warnings", func() {
				Expect(fakeCloudControllerClient.CreateOrganizationQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))

				passedQuota := fakeCloudControllerClient.CreateOrganizationQuotaArgsForCall(0)
				Expect(passedQuota).To(Equal(ccv3Quota))
			})
		})

		When("the create org quota endpoint succeeds", func() {
			var (
				ccv3Quota resources.OrganizationQuota
			)
			BeforeEach(func() {
				ccv3Quota = resources.OrganizationQuota{
					Quota: resources.Quota{
						Name: quotaName,
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
				fakeCloudControllerClient.CreateOrganizationQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("calls the create endpoint with the respective values and returns warnings", func() {
				Expect(fakeCloudControllerClient.CreateOrganizationQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))

				passedQuota := fakeCloudControllerClient.CreateOrganizationQuotaArgsForCall(0)
				Expect(passedQuota).To(Equal(ccv3Quota))
			})
		})
	})

	Describe("UpdateOrganizationQuota", func() {
		var (
			oldQuotaName string
			newQuotaName string
			quotaLimits  QuotaLimits
			warnings     Warnings
			executeErr   error
		)

		BeforeEach(func() {
			oldQuotaName = "old-quota-name"
			newQuotaName = "new-quota-name"
			quotaLimits = QuotaLimits{
				TotalMemoryInMB:       &types.NullInt{Value: 2048, IsSet: true},
				PerProcessMemoryInMB:  &types.NullInt{Value: 1024, IsSet: true},
				TotalInstances:        &types.NullInt{Value: 0, IsSet: false},
				TotalServiceInstances: &types.NullInt{Value: 0, IsSet: true},
				PaidServicesAllowed:   &trueValue,
				TotalRoutes:           &types.NullInt{Value: 6, IsSet: true},
				TotalReservedPorts:    &types.NullInt{Value: 5, IsSet: true},
				TotalLogVolume:        &types.NullInt{Value: 64, IsSet: true},
			}

			fakeCloudControllerClient.GetOrganizationQuotasReturns(
				[]resources.OrganizationQuota{{Quota: resources.Quota{Name: oldQuotaName}}},
				ccv3.Warnings{"get-quotas-warning"},
				nil,
			)
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateOrganizationQuota(oldQuotaName, newQuotaName, quotaLimits)
		})

		When("the update-quota endpoint returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateOrganizationQuotaReturns(
					resources.OrganizationQuota{},
					ccv3.Warnings{"update-quota-warning"},
					errors.New("update-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(fakeCloudControllerClient.UpdateOrganizationQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("get-quotas-warning", "update-quota-warning"))
				Expect(executeErr).To(MatchError("update-error"))
			})
		})

		When("no quota limits are being updated", func() {
			var (
				ccv3Quota resources.OrganizationQuota
			)

			BeforeEach(func() {
				quotaLimits = QuotaLimits{}

				ccv3Quota = resources.OrganizationQuota{
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

				fakeCloudControllerClient.UpdateOrganizationQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"update-quota-warning"},
					nil,
				)
			})

			It("calls the update endpoint with the respective values and returns warnings", func() {
				Expect(fakeCloudControllerClient.UpdateOrganizationQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("get-quotas-warning", "update-quota-warning"))

				passedQuota := fakeCloudControllerClient.UpdateOrganizationQuotaArgsForCall(0)

				updatedQuota := ccv3Quota
				updatedQuota.Name = newQuotaName

				Expect(passedQuota).To(Equal(updatedQuota))
			})
		})

		When("the update org quota has all values set to unlimited", func() {
			var (
				ccv3Quota resources.OrganizationQuota
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
				ccv3Quota = resources.OrganizationQuota{
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
				fakeCloudControllerClient.UpdateOrganizationQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"update-quota-warning"},
					nil,
				)
			})

			It("calls the update endpoint with the respective values and returns warnings", func() {
				Expect(fakeCloudControllerClient.UpdateOrganizationQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("get-quotas-warning", "update-quota-warning"))

				passedQuota := fakeCloudControllerClient.UpdateOrganizationQuotaArgsForCall(0)

				updatedQuota := ccv3Quota
				updatedQuota.Name = newQuotaName

				Expect(passedQuota).To(Equal(updatedQuota))
			})
		})

		When("The update org quota endpoint succeeds", func() {
			var (
				ccv3Quota resources.OrganizationQuota
			)

			BeforeEach(func() {
				ccv3Quota = resources.OrganizationQuota{
					Quota: resources.Quota{
						Name: oldQuotaName,
						Apps: resources.AppLimit{
							TotalMemory:       &types.NullInt{Value: 2048, IsSet: true},
							InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
							TotalAppInstances: &types.NullInt{Value: 0, IsSet: false},
							TotalLogVolume:    &types.NullInt{Value: 64, IsSet: true},
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

				fakeCloudControllerClient.UpdateOrganizationQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"update-quota-warning"},
					nil,
				)
			})

			It("calls the update endpoint with the respective values and returns warnings", func() {
				Expect(fakeCloudControllerClient.UpdateOrganizationQuotaCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("get-quotas-warning", "update-quota-warning"))

				passedQuota := fakeCloudControllerClient.UpdateOrganizationQuotaArgsForCall(0)

				updatedQuota := ccv3Quota
				updatedQuota.Name = newQuotaName

				Expect(passedQuota).To(Equal(updatedQuota))
			})
		})

		When("the org quota name is not being updated", func() {
			var (
				ccv3Quota resources.OrganizationQuota
			)

			BeforeEach(func() {
				newQuotaName = ""

				ccv3Quota = resources.OrganizationQuota{
					Quota: resources.Quota{
						Name: oldQuotaName,
						Apps: resources.AppLimit{
							TotalMemory:       &types.NullInt{Value: 2048, IsSet: true},
							InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
							TotalAppInstances: &types.NullInt{Value: 0, IsSet: false},
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

				fakeCloudControllerClient.UpdateOrganizationQuotaReturns(
					ccv3Quota,
					ccv3.Warnings{"update-quota-warning"},
					nil,
				)
			})
			It("uses the current org quota name in the API request", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				inputQuota := fakeCloudControllerClient.UpdateOrganizationQuotaArgsForCall(0)
				Expect(inputQuota.Name).To(Equal("old-quota-name"))
			})
		})
	})
})
