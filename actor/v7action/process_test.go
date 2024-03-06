package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Process Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("GetProcess", func() {
		var (
			processGUID string

			process  resources.Process
			warnings Warnings
			err      error
		)

		BeforeEach(func() {
			processGUID = "some-process-guid"
		})

		JustBeforeEach(func() {
			process, warnings, err = actor.GetProcess(processGUID)
		})

		When("getting the process is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetProcessReturns(
					resources.Process{
						GUID: "some-process-guid",
					},
					ccv3.Warnings{"some-process-warning"},
					nil,
				)
			})

			It("returns the process and warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-process-warning"))
				Expect(process).To(Equal(resources.Process{
					GUID: "some-process-guid",
				}))

				Expect(fakeCloudControllerClient.GetProcessCallCount()).To(Equal(1))
				passedProcessGUID := fakeCloudControllerClient.GetProcessArgsForCall(0)
				Expect(passedProcessGUID).To(Equal("some-process-guid"))
			})
		})

		When("getting application process by type returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeCloudControllerClient.GetProcessReturns(
					resources.Process{},
					ccv3.Warnings{"some-process-warning"},
					expectedErr,
				)
			})

			It("returns the error and warnings", func() {
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("some-process-warning"))
			})
		})
	})

	Describe("GetProcessByTypeAndApplication", func() {
		var (
			processType string
			appGUID     string

			process  resources.Process
			warnings Warnings
			err      error
		)

		BeforeEach(func() {
			processType = "web"
			appGUID = "some-app-guid"
		})

		JustBeforeEach(func() {
			process, warnings, err = actor.GetProcessByTypeAndApplication(processType, appGUID)
		})

		When("getting the application process is succesful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
					resources.Process{
						GUID: "some-process-guid",
					},
					ccv3.Warnings{"some-process-warning"},
					nil,
				)
			})

			It("returns the process and warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-process-warning"))
				Expect(process).To(Equal(resources.Process{
					GUID: "some-process-guid",
				}))

				Expect(fakeCloudControllerClient.GetApplicationProcessByTypeCallCount()).To(Equal(1))
				passedAppGUID, passedProcessType := fakeCloudControllerClient.GetApplicationProcessByTypeArgsForCall(0)
				Expect(passedAppGUID).To(Equal(appGUID))
				Expect(passedProcessType).To(Equal(processType))
			})
		})

		When("getting application process by type returns an error", func() {
			var expectedErr error

			When("the api returns a ProcessNotFoundError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
						resources.Process{},
						ccv3.Warnings{"some-process-warning"},
						ccerror.ProcessNotFoundError{},
					)
				})

				It("returns a ProcessNotFoundError and all warnings", func() {
					Expect(err).To(Equal(actionerror.ProcessNotFoundError{ProcessType: "web"}))
					Expect(warnings).To(ConsistOf("some-process-warning"))
				})
			})

			Context("generic error", func() {
				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
						resources.Process{},
						ccv3.Warnings{"some-process-warning"},
						expectedErr,
					)
				})

				It("returns the error and warnings", func() {
					Expect(err).To(Equal(expectedErr))
					Expect(warnings).To(ConsistOf("some-process-warning"))
				})
			})
		})
	})

	Describe("ScaleProcessByApplication", func() {
		var (
			passedProcess resources.Process
			warnings      Warnings
			executeErr    error
		)

		BeforeEach(func() {
			passedProcess = resources.Process{
				Type:       constant.ProcessTypeWeb,
				Instances:  types.NullInt{Value: 2, IsSet: true},
				MemoryInMB: types.NullUint64{Value: 100, IsSet: true},
				DiskInMB:   types.NullUint64{Value: 200, IsSet: true},
			}
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.ScaleProcessByApplication("some-app-guid", passedProcess)
		})

		When("no errors are encountered scaling the application process", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationProcessScaleReturns(
					resources.Process{GUID: "some-process-guid"},
					ccv3.Warnings{"scale-process-warning"},
					nil)
			})

			It("scales correct process", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("scale-process-warning"))

				Expect(fakeCloudControllerClient.CreateApplicationProcessScaleCallCount()).To(Equal(1))
				appGUIDArg, processArg := fakeCloudControllerClient.CreateApplicationProcessScaleArgsForCall(0)
				Expect(appGUIDArg).To(Equal("some-app-guid"))
				Expect(processArg).To(Equal(resources.Process{
					Type:       constant.ProcessTypeWeb,
					Instances:  passedProcess.Instances,
					MemoryInMB: passedProcess.MemoryInMB,
					DiskInMB:   passedProcess.DiskInMB,
				}))
			})
		})

		When("an error is encountered scaling the application process", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("scale process error")
				fakeCloudControllerClient.CreateApplicationProcessScaleReturns(
					resources.Process{GUID: "some-process-guid"},
					ccv3.Warnings{"scale-process-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("scale-process-warning"))
			})
		})

		When("a ProcessNotFoundError error is encountered scaling the application process", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationProcessScaleReturns(
					resources.Process{GUID: "some-process-guid"},
					ccv3.Warnings{"scale-process-warning"},
					ccerror.ProcessNotFoundError{},
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(Equal(actionerror.ProcessNotFoundError{ProcessType: constant.ProcessTypeWeb}))
				Expect(warnings).To(ConsistOf("scale-process-warning"))
			})
		})
	})

	Describe("UpdateProcessByTypeAndApplication", func() {
		var (
			processType  string
			appGUID      string
			inputProcess resources.Process

			warnings Warnings
			err      error
		)

		BeforeEach(func() {
			processType = "web"
			appGUID = "some-app-guid"
			inputProcess = resources.Process{}
		})

		JustBeforeEach(func() {
			warnings, err = actor.UpdateProcessByTypeAndApplication(processType, appGUID, inputProcess)
		})

		When("the user specifies an endpoint for a non-http health check", func() {
			BeforeEach(func() {
				inputProcess.HealthCheckType = constant.Port
				inputProcess.HealthCheckEndpoint = "some-http-endpoint"
			})

			It("returns an HTTPHealthCheckInvalidError", func() {
				Expect(err).To(MatchError(actionerror.HTTPHealthCheckInvalidError{}))
				Expect(warnings).To(BeNil())
			})
		})

		When("getting application process by type returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
					resources.Process{},
					ccv3.Warnings{"some-process-warning"},
					expectedErr,
				)
			})

			It("returns the error and warnings", func() {
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("some-process-warning"))
			})
		})

		When("application process exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
					resources.Process{
						GUID: "some-process-guid",
					},
					ccv3.Warnings{"some-process-warning"},
					nil,
				)
			})

			When("updating the process errors", func() {
				var expectedErr error

				BeforeEach(func() {
					inputProcess.HealthCheckType = constant.Port
					inputProcess.HealthCheckEndpoint = constant.ProcessHealthCheckEndpointDefault

					expectedErr = errors.New("some-error")
					fakeCloudControllerClient.UpdateProcessReturns(
						resources.Process{},
						ccv3.Warnings{"some-health-check-warning"},
						expectedErr,
					)
				})

				It("returns the error and warnings", func() {
					Expect(err).To(Equal(expectedErr))
					Expect(warnings).To(ConsistOf("some-process-warning", "some-health-check-warning"))
				})
			})

			When("update the process is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateProcessReturns(
						resources.Process{GUID: "some-process-guid"},
						ccv3.Warnings{"some-health-check-warning"},
						nil,
					)
				})

				It("gets the correct application process", func() {
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeCloudControllerClient.GetApplicationProcessByTypeCallCount()).To(Equal(1))
					passedAppGUID, passedProcessType := fakeCloudControllerClient.GetApplicationProcessByTypeArgsForCall(0)
					Expect(passedAppGUID).To(Equal(appGUID))
					Expect(passedProcessType).To(Equal(processType))
				})

				When("updating the command is successful", func() {
					var command types.FilteredString
					BeforeEach(func() {
						command = *types.NewFilteredString("some-command")
						inputProcess.Command = command
					})

					It("returns the application", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(warnings).To(ConsistOf("some-process-warning", "some-health-check-warning"))

						Expect(fakeCloudControllerClient.UpdateProcessCallCount()).To(Equal(1))
						process := fakeCloudControllerClient.UpdateProcessArgsForCall(0)
						Expect(process).To(MatchFields(IgnoreExtras,
							Fields{
								"GUID":    Equal("some-process-guid"),
								"Command": Equal(command),
							}))
					})
				})

				When("the health check type is http", func() {
					BeforeEach(func() {
						inputProcess.HealthCheckType = constant.HTTP
						inputProcess.HealthCheckEndpoint = "some-http-endpoint"
						inputProcess.HealthCheckInvocationTimeout = 42
					})

					It("returns the application", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(warnings).To(ConsistOf("some-process-warning", "some-health-check-warning"))

						Expect(fakeCloudControllerClient.UpdateProcessCallCount()).To(Equal(1))
						process := fakeCloudControllerClient.UpdateProcessArgsForCall(0)
						Expect(process).To(MatchFields(IgnoreExtras,
							Fields{
								"GUID":                         Equal("some-process-guid"),
								"HealthCheckType":              Equal(constant.HTTP),
								"HealthCheckEndpoint":          Equal("some-http-endpoint"),
								"HealthCheckInvocationTimeout": BeEquivalentTo(42),
							}))
					})
				})

				When("the health check type is not http", func() {
					BeforeEach(func() {
						inputProcess.HealthCheckType = constant.Port
						inputProcess.HealthCheckInvocationTimeout = 42
					})

					When("the endpoint is set to '/'", func() {
						BeforeEach(func() {
							inputProcess.HealthCheckEndpoint = constant.ProcessHealthCheckEndpointDefault
						})

						It("does not send the / endpoint and returns the application", func() {
							Expect(err).NotTo(HaveOccurred())
							Expect(warnings).To(ConsistOf("some-process-warning", "some-health-check-warning"))

							Expect(fakeCloudControllerClient.UpdateProcessCallCount()).To(Equal(1))
							process := fakeCloudControllerClient.UpdateProcessArgsForCall(0)
							Expect(process).To(MatchFields(IgnoreExtras,
								Fields{
									"GUID":                         Equal("some-process-guid"),
									"HealthCheckType":              Equal(constant.Port),
									"HealthCheckEndpoint":          BeEmpty(),
									"HealthCheckInvocationTimeout": BeEquivalentTo(42),
								}))
						})
					})

					When("the endpoint is empty", func() {
						It("does not send the / endpoint and returns the application", func() {
							Expect(err).NotTo(HaveOccurred())
							Expect(warnings).To(ConsistOf("some-process-warning", "some-health-check-warning"))

							Expect(fakeCloudControllerClient.UpdateProcessCallCount()).To(Equal(1))
							process := fakeCloudControllerClient.UpdateProcessArgsForCall(0)
							Expect(process).To(MatchFields(IgnoreExtras,
								Fields{
									"GUID":                         Equal("some-process-guid"),
									"HealthCheckType":              Equal(constant.Port),
									"HealthCheckEndpoint":          BeEmpty(),
									"HealthCheckInvocationTimeout": BeEquivalentTo(42),
								}))
						})
					})
				})
			})
		})
	})
})
