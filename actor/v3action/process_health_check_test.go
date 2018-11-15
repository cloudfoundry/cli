package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Process Health Check Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("ProcessHealthChecks", func() {
		var healthchecks ProcessHealthChecks

		BeforeEach(func() {
			healthchecks = ProcessHealthChecks{
				{
					ProcessType:     "worker",
					HealthCheckType: "process",
				},
				{
					ProcessType:     "console",
					HealthCheckType: "process",
				},
				{
					ProcessType:     constant.ProcessTypeWeb,
					HealthCheckType: "http",
					Endpoint:        "/",
				},
			}
		})

		Describe("Sort", func() {
			It("sorts healthchecks with web first and then alphabetically sorted", func() {
				healthchecks.Sort()
				Expect(healthchecks[0].ProcessType).To(Equal(constant.ProcessTypeWeb))
				Expect(healthchecks[1].ProcessType).To(Equal("console"))
				Expect(healthchecks[2].ProcessType).To(Equal("worker"))
			})
		})
	})

	Describe("GetApplicationProcessHealthChecksByNameAndSpace", func() {
		var (
			warnings            Warnings
			executeErr          error
			processHealthChecks []ProcessHealthCheck
		)

		JustBeforeEach(func() {
			processHealthChecks, warnings, executeErr = actor.GetApplicationProcessHealthChecksByNameAndSpace("some-app-name", "some-space-guid")
		})

		When("application does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
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
					[]ccv3.Application{},
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
					[]ccv3.Application{
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
						[]ccv3.Process{},
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
						[]ccv3.Process{
							{
								GUID:                         "process-guid-1",
								Type:                         "process-type-1",
								HealthCheckType:              "health-check-type-1",
								HealthCheckEndpoint:          "health-check-endpoint-1",
								HealthCheckInvocationTimeout: 42,
							},
							{
								GUID:                         "process-guid-2",
								Type:                         "process-type-2",
								HealthCheckType:              "health-check-type-2",
								HealthCheckInvocationTimeout: 0,
							},
						},
						ccv3.Warnings{"some-process-warning"},
						nil,
					)
				})

				It("returns health checks", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-warning", "some-process-warning"))
					Expect(processHealthChecks).To(Equal([]ProcessHealthCheck{
						{
							ProcessType:       "process-type-1",
							HealthCheckType:   "health-check-type-1",
							Endpoint:          "health-check-endpoint-1",
							InvocationTimeout: 42,
						},
						{
							ProcessType:       "process-type-2",
							HealthCheckType:   "health-check-type-2",
							InvocationTimeout: 0,
						},
					}))
				})
			})
		})
	})

	Describe("SetApplicationProcessHealthCheckTypeByNameAndSpace", func() {
		var (
			healthCheckType     string
			healthCheckEndpoint string

			warnings Warnings
			err      error
			app      Application
		)

		BeforeEach(func() {
			healthCheckType = "port"
			healthCheckEndpoint = "/"
		})

		JustBeforeEach(func() {
			app, warnings, err = actor.SetApplicationProcessHealthCheckTypeByNameAndSpace("some-app-name", "some-space-guid", healthCheckType, healthCheckEndpoint, "some-process-type", 42)
		})

		When("the user specifies an endpoint for a non-http health check", func() {
			BeforeEach(func() {
				healthCheckType = "port"
				healthCheckEndpoint = "some-http-endpoint"
			})

			It("returns an HTTPHealthCheckInvalidError", func() {
				Expect(err).To(MatchError(actionerror.HTTPHealthCheckInvalidError{}))
				Expect(warnings).To(BeNil())
			})
		})

		When("application does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					nil,
				)

				healthCheckType = "http"
				healthCheckEndpoint = "some-http-endpoint"
			})

			It("returns the error and warnings", func() {
				Expect(err).To(Equal(actionerror.ApplicationNotFoundError{Name: "some-app-name"}))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})

		When("getting application returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					expectedErr,
				)
			})

			It("returns the error and warnings", func() {
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})

		When("application exists", func() {
			var ccv3App ccv3.Application

			BeforeEach(func() {
				ccv3App = ccv3.Application{
					GUID: "some-app-guid",
				}

				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{ccv3App},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			When("getting application process by type returns an error", func() {
				var expectedErr error

				When("the api returns a ProcessNotFoundError", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
							ccv3.Process{},
							ccv3.Warnings{"some-process-warning"},
							ccerror.ProcessNotFoundError{},
						)
					})

					It("returns a ProcessNotFoundError and all warnings", func() {
						Expect(err).To(Equal(actionerror.ProcessNotFoundError{ProcessType: "some-process-type"}))
						Expect(warnings).To(ConsistOf("some-warning", "some-process-warning"))
					})
				})

				Context("generic error", func() {
					BeforeEach(func() {
						expectedErr = errors.New("some-error")
						fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
							ccv3.Process{},
							ccv3.Warnings{"some-process-warning"},
							expectedErr,
						)
					})

					It("returns the error and warnings", func() {
						Expect(err).To(Equal(expectedErr))
						Expect(warnings).To(ConsistOf("some-warning", "some-process-warning"))
					})
				})
			})

			When("application process exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
						ccv3.Process{
							GUID: "some-process-guid",
						},
						ccv3.Warnings{"some-process-warning"},
						nil,
					)
				})

				When("setting process health check type returns an error", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("some-error")
						fakeCloudControllerClient.UpdateProcessReturns(
							ccv3.Process{},
							ccv3.Warnings{"some-health-check-warning"},
							expectedErr,
						)
					})

					It("returns the error and warnings", func() {
						Expect(err).To(Equal(expectedErr))
						Expect(warnings).To(ConsistOf("some-warning", "some-process-warning", "some-health-check-warning"))
					})
				})

				When("setting process health check type succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.UpdateProcessReturns(
							ccv3.Process{GUID: "some-process-guid"},
							ccv3.Warnings{"some-health-check-warning"},
							nil,
						)
					})

					When("the health check type is http", func() {
						BeforeEach(func() {
							healthCheckType = "http"
							healthCheckEndpoint = "some-http-endpoint"
						})

						It("returns the application", func() {
							Expect(err).NotTo(HaveOccurred())
							Expect(warnings).To(ConsistOf("some-warning", "some-process-warning", "some-health-check-warning"))

							Expect(app).To(Equal(Application{
								GUID: ccv3App.GUID,
							}))

							Expect(fakeCloudControllerClient.GetApplicationProcessByTypeCallCount()).To(Equal(1))
							appGUID, processType := fakeCloudControllerClient.GetApplicationProcessByTypeArgsForCall(0)
							Expect(appGUID).To(Equal("some-app-guid"))
							Expect(processType).To(Equal("some-process-type"))

							Expect(fakeCloudControllerClient.UpdateProcessCallCount()).To(Equal(1))
							process := fakeCloudControllerClient.UpdateProcessArgsForCall(0)
							Expect(process.GUID).To(Equal("some-process-guid"))
							Expect(process.HealthCheckType).To(Equal("http"))
							Expect(process.HealthCheckEndpoint).To(Equal("some-http-endpoint"))
							Expect(process.HealthCheckInvocationTimeout).To(Equal(42))
						})
					})

					When("the health check type is not http", func() {
						It("does not send the / endpoint and returns the application", func() {
							Expect(err).NotTo(HaveOccurred())
							Expect(warnings).To(ConsistOf("some-warning", "some-process-warning", "some-health-check-warning"))

							Expect(app).To(Equal(Application{
								GUID: ccv3App.GUID,
							}))

							Expect(fakeCloudControllerClient.GetApplicationProcessByTypeCallCount()).To(Equal(1))
							appGUID, processType := fakeCloudControllerClient.GetApplicationProcessByTypeArgsForCall(0)
							Expect(appGUID).To(Equal("some-app-guid"))
							Expect(processType).To(Equal("some-process-type"))

							Expect(fakeCloudControllerClient.UpdateProcessCallCount()).To(Equal(1))
							process := fakeCloudControllerClient.UpdateProcessArgsForCall(0)
							Expect(process.GUID).To(Equal("some-process-guid"))
							Expect(process.HealthCheckType).To(Equal("port"))
							Expect(process.HealthCheckEndpoint).To(BeEmpty())
							Expect(process.HealthCheckInvocationTimeout).To(Equal(42))
						})
					})
				})
			})
		})
	})
})
