package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	. "code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Process Readiness Health Check Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("ProcessReadinessHealthChecks", func() {
		var readinessHealthChecks ProcessReadinessHealthChecks

		BeforeEach(func() {
			readinessHealthChecks = ProcessReadinessHealthChecks{
				{
					ProcessType:     "worker",
					HealthCheckType: constant.Process,
				},
				{
					ProcessType:     "console",
					HealthCheckType: constant.Process,
				},
				{
					ProcessType:     constant.ProcessTypeWeb,
					HealthCheckType: constant.HTTP,
					Endpoint:        constant.ProcessHealthCheckEndpointDefault,
				},
			}
		})

		Describe("Sort", func() {
			It("sorts readiness health checks with web first and then alphabetically sorted", func() {
				readinessHealthChecks.Sort()
				Expect(readinessHealthChecks[0].ProcessType).To(Equal(constant.ProcessTypeWeb))
				Expect(readinessHealthChecks[1].ProcessType).To(Equal("console"))
				Expect(readinessHealthChecks[2].ProcessType).To(Equal("worker"))
			})
		})
	})

	Describe("GetApplicationProcessReadinessHealthChecksByNameAndSpace", func() {
		var (
			warnings                     Warnings
			executeErr                   error
			processReadinessHealthChecks []ProcessReadinessHealthCheck
		)

		JustBeforeEach(func() {
			processReadinessHealthChecks, warnings, executeErr = actor.GetApplicationProcessReadinessHealthChecksByNameAndSpace("some-app-name", "some-space-guid")
		})

		When("application does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: "some-app-name"}))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})

		When("getting application returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{},
					ccv3.Warnings{"some-warning"},
					expectedErr,
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})

		When("application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							GUID: "some-app-guid",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			When("getting application processes returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeCloudControllerClient.GetApplicationProcessesReturns(
						[]resources.Process{},
						ccv3.Warnings{"some-process-warning"},
						expectedErr,
					)
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(warnings).To(ConsistOf("some-warning", "some-process-warning"))
				})
			})

			When("application has processes", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationProcessesReturns(
						[]resources.Process{
							{
								GUID:                                  "process-guid-1",
								Type:                                  "process-type-1",
								ReadinessHealthCheckType:              "readiness-health-check-type-1",
								ReadinessHealthCheckEndpoint:          "readiness-health-check-endpoint-1",
								ReadinessHealthCheckInvocationTimeout: 42,
							},
							{
								GUID:                                  "process-guid-2",
								Type:                                  "process-type-2",
								ReadinessHealthCheckType:              "readiness-health-check-type-2",
								ReadinessHealthCheckInvocationTimeout: 0,
							},
						},
						ccv3.Warnings{"some-process-warning"},
						nil,
					)
				})

				It("returns health checks", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-warning", "some-process-warning"))
					Expect(processReadinessHealthChecks).To(Equal([]ProcessReadinessHealthCheck{
						{
							ProcessType:       "process-type-1",
							HealthCheckType:   "readiness-health-check-type-1",
							Endpoint:          "readiness-health-check-endpoint-1",
							InvocationTimeout: 42,
						},
						{
							ProcessType:       "process-type-2",
							HealthCheckType:   "readiness-health-check-type-2",
							InvocationTimeout: 0,
						},
					}))
				})
			})
		})
	})
})
