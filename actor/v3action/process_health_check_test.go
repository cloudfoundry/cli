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
		Context("when application does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetApplicationProcessHealthChecksByNameAndSpace("some-app-name", "some-space-guid")
				Expect(err).To(Equal(actionerror.ApplicationNotFoundError{Name: "some-app-name"}))
				Expect(warnings).To(Equal(Warnings{"some-warning"}))
			})
		})

		Context("when getting application returns an error", func() {
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
				_, warnings, err := actor.GetApplicationProcessHealthChecksByNameAndSpace("some-app-name", "some-space-guid")
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(Equal(Warnings{"some-warning"}))
			})
		})

		Context("when application exists", func() {
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

			Context("when getting application processes returns an error", func() {
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
					_, warnings, err := actor.GetApplicationProcessHealthChecksByNameAndSpace("some-app-name", "some-space-guid")
					Expect(err).To(Equal(expectedErr))
					Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning"}))
				})
			})

			Context("when application has processes", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationProcessesReturns(
						[]ccv3.Process{
							{
								GUID: "process-guid-1",
								Type: "process-type-1",
								HealthCheck: ccv3.ProcessHealthCheck{
									Type: "health-check-type-1",
									Data: ccv3.ProcessHealthCheckData{
										Endpoint: "health-check-endpoint-1",
									},
								},
							},
							{
								GUID: "process-guid-2",
								Type: "process-type-2",
								HealthCheck: ccv3.ProcessHealthCheck{
									Type: "health-check-type-2",
									Data: ccv3.ProcessHealthCheckData{},
								},
							},
						},
						ccv3.Warnings{"some-process-warning"},
						nil,
					)
				})

				It("returns health checks", func() {
					processHealthChecks, warnings, err := actor.GetApplicationProcessHealthChecksByNameAndSpace("some-app-name", "some-space-guid")
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning"}))
					Expect(processHealthChecks).To(Equal([]ProcessHealthCheck{
						{
							ProcessType:     "process-type-1",
							HealthCheckType: "health-check-type-1",
							Endpoint:        "health-check-endpoint-1",
						},
						{
							ProcessType:     "process-type-2",
							HealthCheckType: "health-check-type-2",
						},
					}))
				})
			})
		})
	})

	Describe("SetApplicationProcessHealthCheckTypeByNameAndSpace", func() {
		Context("when the user specifies an endpoint for a non-http health check", func() {
			It("returns an HTTPHealthCheckInvalidError", func() {
				_, warnings, err := actor.SetApplicationProcessHealthCheckTypeByNameAndSpace("some-app-name", "some-space-guid", "port", "some-http-endpoint", "some-process-type")
				Expect(err).To(MatchError(actionerror.HTTPHealthCheckInvalidError{}))
				Expect(warnings).To(BeNil())
			})
		})

		Context("when application does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.SetApplicationProcessHealthCheckTypeByNameAndSpace("some-app-name", "some-space-guid", "http", "some-http-endpoint", "some-process-type")
				Expect(err).To(Equal(actionerror.ApplicationNotFoundError{Name: "some-app-name"}))
				Expect(warnings).To(Equal(Warnings{"some-warning"}))
			})
		})

		Context("when getting application returns an error", func() {
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
				_, warnings, err := actor.SetApplicationProcessHealthCheckTypeByNameAndSpace("some-app-name", "some-space-guid", "http", "some-http-endpoint", "some-process-type")
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(Equal(Warnings{"some-warning"}))
			})
		})

		Context("when application exists", func() {
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

			Context("when getting application process by type returns an error", func() {
				var expectedErr error

				Context("when the api returns a ProcessNotFoundError", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
							ccv3.Process{},
							ccv3.Warnings{"some-process-warning"},
							ccerror.ProcessNotFoundError{},
						)
					})

					It("returns a ProcessNotFoundError and all warnings", func() {
						_, warnings, err := actor.SetApplicationProcessHealthCheckTypeByNameAndSpace("some-app-name", "some-space-guid", "http", "some-http-endpoint", "some-process-type")
						Expect(err).To(Equal(actionerror.ProcessNotFoundError{ProcessType: "some-process-type"}))
						Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning"}))
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
						_, warnings, err := actor.SetApplicationProcessHealthCheckTypeByNameAndSpace("some-app-name", "some-space-guid", "http", "some-http-endpoint", "some-process-type")
						Expect(err).To(Equal(expectedErr))
						Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning"}))
					})
				})
			})

			Context("when application process exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
						ccv3.Process{
							GUID: "some-process-guid",
						},
						ccv3.Warnings{"some-process-warning"},
						nil,
					)
				})

				Context("when setting process health check type returns an error", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("some-error")
						fakeCloudControllerClient.PatchApplicationProcessHealthCheckReturns(
							ccv3.Process{},
							ccv3.Warnings{"some-health-check-warning"},
							expectedErr,
						)
					})

					It("returns the error and warnings", func() {
						_, warnings, err := actor.SetApplicationProcessHealthCheckTypeByNameAndSpace("some-app-name", "some-space-guid", "http", "some-http-endpoint", "some-process-type")
						Expect(err).To(Equal(expectedErr))
						Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning", "some-health-check-warning"}))
					})
				})

				Context("when setting process health check type succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.PatchApplicationProcessHealthCheckReturns(
							ccv3.Process{GUID: "some-process-guid"},
							ccv3.Warnings{"some-health-check-warning"},
							nil,
						)
					})
					Context("when the health check type is http", func() {
						It("returns the application", func() {
							app, warnings, err := actor.SetApplicationProcessHealthCheckTypeByNameAndSpace("some-app-name", "some-space-guid", "http", "some-http-endpoint", "some-process-type")
							Expect(err).NotTo(HaveOccurred())
							Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning", "some-health-check-warning"}))

							Expect(app).To(Equal(Application{
								GUID: ccv3App.GUID,
							}))

							Expect(fakeCloudControllerClient.GetApplicationProcessByTypeCallCount()).To(Equal(1))
							appGUID, processType := fakeCloudControllerClient.GetApplicationProcessByTypeArgsForCall(0)
							Expect(appGUID).To(Equal("some-app-guid"))
							Expect(processType).To(Equal("some-process-type"))

							Expect(fakeCloudControllerClient.PatchApplicationProcessHealthCheckCallCount()).To(Equal(1))
							processGUID, processHealthCheckType, processHealthCheckEndpoint := fakeCloudControllerClient.PatchApplicationProcessHealthCheckArgsForCall(0)
							Expect(processGUID).To(Equal("some-process-guid"))
							Expect(processHealthCheckType).To(Equal("http"))
							Expect(processHealthCheckEndpoint).To(Equal("some-http-endpoint"))
						})
					})
					Context("when the health check type is not http", func() {
						It("does not send the / endpoint and returns the application", func() {
							app, warnings, err := actor.SetApplicationProcessHealthCheckTypeByNameAndSpace("some-app-name", "some-space-guid", "port", "/", "some-process-type")
							Expect(err).NotTo(HaveOccurred())
							Expect(warnings).To(Equal(Warnings{"some-warning", "some-process-warning", "some-health-check-warning"}))

							Expect(app).To(Equal(Application{
								GUID: ccv3App.GUID,
							}))

							Expect(fakeCloudControllerClient.GetApplicationProcessByTypeCallCount()).To(Equal(1))
							appGUID, processType := fakeCloudControllerClient.GetApplicationProcessByTypeArgsForCall(0)
							Expect(appGUID).To(Equal("some-app-guid"))
							Expect(processType).To(Equal("some-process-type"))

							Expect(fakeCloudControllerClient.PatchApplicationProcessHealthCheckCallCount()).To(Equal(1))
							processGUID, processHealthCheckType, processHealthCheckEndpoint := fakeCloudControllerClient.PatchApplicationProcessHealthCheckArgsForCall(0)
							Expect(processGUID).To(Equal("some-process-guid"))
							Expect(processHealthCheckType).To(Equal("port"))
							Expect(processHealthCheckEndpoint).To(BeEmpty())
						})
					})
				})
			})
		})
	})
})
