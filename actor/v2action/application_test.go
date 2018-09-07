package v2action_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/types"

	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		fakeConfig                *v2actionfakes.FakeConfig
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, fakeConfig = NewTestActor()
		fakeConfig.DialTimeoutReturns(time.Millisecond)
	})

	Describe("Application", func() {
		var app Application
		BeforeEach(func() {
			app = Application{}
		})

		Describe("CalculatedCommand", func() {
			When("command is set", func() {
				BeforeEach(func() {
					app.Command = types.FilteredString{IsSet: true, Value: "foo"}
					app.DetectedStartCommand = types.FilteredString{IsSet: true, Value: "bar"}
				})

				It("returns back the command", func() {
					Expect(app.CalculatedCommand()).To(Equal("foo"))
				})
			})

			Context("only detected start command is set", func() {
				BeforeEach(func() {
					app.DetectedStartCommand = types.FilteredString{IsSet: true, Value: "bar"}
				})

				It("returns back the detected start command", func() {
					Expect(app.CalculatedCommand()).To(Equal("bar"))
				})
			})

			Context("neither command nor detected start command are set", func() {
				It("returns an empty string", func() {
					Expect(app.CalculatedCommand()).To(BeEmpty())
				})
			})
		})

		Describe("CalculatedBuildpack", func() {
			When("buildpack is set", func() {
				BeforeEach(func() {
					app.Buildpack = types.FilteredString{IsSet: true, Value: "foo"}
					app.DetectedBuildpack = types.FilteredString{IsSet: true, Value: "bar"}
				})

				It("returns back the buildpack", func() {
					Expect(app.CalculatedBuildpack()).To(Equal("foo"))
				})
			})

			Context("only detected buildpack is set", func() {
				BeforeEach(func() {
					app.DetectedBuildpack = types.FilteredString{IsSet: true, Value: "bar"}
				})

				It("returns back the detected buildpack", func() {
					Expect(app.CalculatedBuildpack()).To(Equal("bar"))
				})
			})

			Context("neither buildpack nor detected buildpack are set", func() {
				It("returns an empty string", func() {
					Expect(app.CalculatedBuildpack()).To(BeEmpty())
				})
			})
		})

		Describe("CalculatedHealthCheckEndpoint", func() {
			When("the health check type is http", func() {
				BeforeEach(func() {
					app.HealthCheckType = "http"
					app.HealthCheckHTTPEndpoint = "/some-endpoint"
				})

				It("returns the endpoint field", func() {
					Expect(app.CalculatedHealthCheckEndpoint()).To(Equal(
						"/some-endpoint"))
				})
			})

			When("the health check type is not http", func() {
				BeforeEach(func() {
					app.HealthCheckType = "process"
					app.HealthCheckHTTPEndpoint = "/some-endpoint"
				})

				It("returns the empty string", func() {
					Expect(app.CalculatedHealthCheckEndpoint()).To(Equal(""))
				})
			})
		})

		Describe("StagingCompleted", func() {
			When("staging the application completes", func() {
				It("returns true", func() {
					app.PackageState = constant.ApplicationPackageStaged
					Expect(app.StagingCompleted()).To(BeTrue())
				})
			})

			When("the application is *not* staged", func() {
				It("returns false", func() {
					app.PackageState = constant.ApplicationPackageFailed
					Expect(app.StagingCompleted()).To(BeFalse())
				})
			})
		})

		Describe("StagingFailed", func() {
			When("staging the application fails", func() {
				It("returns true", func() {
					app.PackageState = constant.ApplicationPackageFailed
					Expect(app.StagingFailed()).To(BeTrue())
				})
			})

			When("staging the application does *not* fail", func() {
				It("returns false", func() {
					app.PackageState = constant.ApplicationPackageStaged
					Expect(app.StagingFailed()).To(BeFalse())
				})
			})
		})

		Describe("StagingFailedMessage", func() {
			When("the application has a staging failed description", func() {
				BeforeEach(func() {
					app.StagingFailedDescription = "An app was not successfully detected by any available buildpack"
					app.StagingFailedReason = "NoAppDetectedError"
				})
				It("returns that description", func() {
					Expect(app.StagingFailedMessage()).To(Equal("An app was not successfully detected by any available buildpack"))
				})
			})

			When("the application does not have a staging failed description", func() {
				BeforeEach(func() {
					app.StagingFailedDescription = ""
					app.StagingFailedReason = "NoAppDetectedError"
				})
				It("returns the staging failed code", func() {
					Expect(app.StagingFailedMessage()).To(Equal("NoAppDetectedError"))
				})
			})
		})

		Describe("StagingFailedNoAppDetected", func() {
			When("staging the application fails due to a no app detected error", func() {
				It("returns true", func() {
					app.StagingFailedReason = "NoAppDetectedError"
					Expect(app.StagingFailedNoAppDetected()).To(BeTrue())
				})
			})

			When("staging the application fails due to any other reason", func() {
				It("returns false", func() {
					app.StagingFailedReason = "InsufficientResources"
					Expect(app.StagingFailedNoAppDetected()).To(BeFalse())
				})
			})
		})

		Describe("Started", func() {
			When("app is started", func() {
				It("returns true", func() {
					Expect(Application{State: constant.ApplicationStarted}.Started()).To(BeTrue())
				})
			})

			When("app is stopped", func() {
				It("returns false", func() {
					Expect(Application{State: constant.ApplicationStopped}.Started()).To(BeFalse())
				})
			})
		})

		Describe("Stopped", func() {
			When("app is started", func() {
				It("returns true", func() {
					Expect(Application{State: constant.ApplicationStopped}.Stopped()).To(BeTrue())
				})
			})

			When("app is stopped", func() {
				It("returns false", func() {
					Expect(Application{State: constant.ApplicationStarted}.Stopped()).To(BeFalse())
				})
			})
		})
	})

	Describe("CreateApplication", func() {
		When("the create is successful", func() {
			var expectedApp ccv2.Application
			BeforeEach(func() {
				expectedApp = ccv2.Application{
					GUID:      "some-app-guid",
					Name:      "some-app-name",
					SpaceGUID: "some-space-guid",
				}
				fakeCloudControllerClient.CreateApplicationReturns(expectedApp, ccv2.Warnings{"some-app-warning-1"}, nil)
			})

			It("creates and returns the application", func() {
				newApp := Application{
					Name:      "some-app-name",
					SpaceGUID: "some-space-guid",
				}
				app, warnings, err := actor.CreateApplication(newApp)
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-app-warning-1"))
				Expect(app).To(Equal(Application(expectedApp)))

				Expect(fakeCloudControllerClient.CreateApplicationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateApplicationArgsForCall(0)).To(Equal(ccv2.Application(newApp)))
			})
		})

		When("the client returns back an error", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some create app error")
				fakeCloudControllerClient.CreateApplicationReturns(ccv2.Application{}, ccv2.Warnings{"some-app-warning-1"}, expectedErr)
			})

			It("returns warnings and an error", func() {
				newApp := Application{
					Name:      "some-app-name",
					SpaceGUID: "some-space-guid",
				}
				_, warnings, err := actor.CreateApplication(newApp)
				Expect(warnings).To(ConsistOf("some-app-warning-1"))
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})

	Describe("GetApplication", func() {
		When("the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationReturns(
					ccv2.Application{
						GUID: "some-app-guid",
						Name: "some-app",
					},
					ccv2.Warnings{"foo"},
					nil,
				)
			})

			It("returns the application and warnings", func() {
				app, warnings, err := actor.GetApplication("some-app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(app).To(Equal(Application{
					GUID: "some-app-guid",
					Name: "some-app",
				}))
				Expect(warnings).To(Equal(Warnings{"foo"}))

				Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		When("the application does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationReturns(ccv2.Application{}, nil, ccerror.ResourceNotFoundError{})
			})

			It("returns an ApplicationNotFoundError", func() {
				_, _, err := actor.GetApplication("some-app-guid")
				Expect(err).To(MatchError(actionerror.ApplicationNotFoundError{GUID: "some-app-guid"}))
			})
		})
	})

	Describe("GetApplicationByNameAndSpace", func() {
		When("the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{
						{
							GUID: "some-app-guid",
							Name: "some-app",
						},
					},
					ccv2.Warnings{"foo"},
					nil,
				)
			})

			It("returns the application and warnings", func() {
				app, warnings, err := actor.GetApplicationByNameAndSpace("some-app", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(app).To(Equal(Application{
					GUID: "some-app-guid",
					Name: "some-app",
				}))
				Expect(warnings).To(Equal(Warnings{"foo"}))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf([]ccv2.Filter{
					ccv2.Filter{
						Type:     constant.NameFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"some-app"},
					},
					ccv2.Filter{
						Type:     constant.SpaceGUIDFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"some-space-guid"},
					},
				}))
			})
		})

		When("the application does not exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv2.Application{}, nil, nil)
			})

			It("returns an ApplicationNotFoundError", func() {
				_, _, err := actor.GetApplicationByNameAndSpace("some-app", "some-space-guid")
				Expect(err).To(MatchError(actionerror.ApplicationNotFoundError{Name: "some-app"}))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns([]ccv2.Application{}, nil, expectedError)
			})

			It("returns the error", func() {
				_, _, err := actor.GetApplicationByNameAndSpace("some-app", "some-space-guid")
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("GetApplicationsBySpace", func() {
		When("the there are applications in the space", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{
						{
							GUID: "some-app-guid-1",
							Name: "some-app-1",
						},
						{
							GUID: "some-app-guid-2",
							Name: "some-app-2",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns the application and warnings", func() {
				apps, warnings, err := actor.GetApplicationsBySpace("some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(apps).To(ConsistOf(
					Application{
						GUID: "some-app-guid-1",
						Name: "some-app-1",
					},
					Application{
						GUID: "some-app-guid-2",
						Name: "some-app-2",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf([]ccv2.Filter{
					ccv2.Filter{
						Type:     constant.SpaceGUIDFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"some-space-guid"},
					},
				}))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("some cc error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{},
					ccv2.Warnings{"warning-1", "warning-2"},
					expectedError)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetApplicationsBySpace("some-space-guid")
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("GetRouteApplications", func() {
		When("the CC client returns no errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRouteApplicationsReturns(
					[]ccv2.Application{
						{
							GUID: "application-guid",
							Name: "application-name",
						},
					}, ccv2.Warnings{"route-applications-warning"}, nil)
			})
			It("returns the applications bound to the route and warnings", func() {
				applications, warnings, err := actor.GetRouteApplications("route-guid")
				Expect(fakeCloudControllerClient.GetRouteApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRouteApplicationsArgsForCall(0)).To(Equal("route-guid"))

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("route-applications-warning"))
				Expect(applications).To(ConsistOf(
					Application{
						GUID: "application-guid",
						Name: "application-name",
					},
				))
			})
		})

		When("the CC client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRouteApplicationsReturns(
					[]ccv2.Application{}, ccv2.Warnings{"route-applications-warning"}, errors.New("get-route-applications-error"))
			})

			It("returns the error and warnings", func() {
				apps, warnings, err := actor.GetRouteApplications("route-guid")
				Expect(fakeCloudControllerClient.GetRouteApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRouteApplicationsArgsForCall(0)).To(Equal("route-guid"))

				Expect(err).To(MatchError("get-route-applications-error"))
				Expect(warnings).To(ConsistOf("route-applications-warning"))
				Expect(apps).To(BeNil())
			})
		})
	})

	Describe("SetApplicationHealthCheckTypeByNameAndSpace", func() {
		When("setting an http endpoint with a health check that is not http", func() {
			It("returns an http health check invalid error", func() {
				_, _, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
					"some-app", "some-space-guid", "some-health-check-type", "/foo")
				Expect(err).To(MatchError(actionerror.HTTPHealthCheckInvalidError{}))
			})
		})

		When("the app exists", func() {
			When("the desired health check type is different", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{
							{GUID: "some-app-guid"},
						},
						ccv2.Warnings{"get application warning"},
						nil,
					)
					fakeCloudControllerClient.UpdateApplicationReturns(
						ccv2.Application{
							GUID:            "some-app-guid",
							HealthCheckType: "process",
						},
						ccv2.Warnings{"update warnings"},
						nil,
					)
				})

				It("sets the desired health check type and returns the warnings", func() {
					returnedApp, warnings, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
						"some-app", "some-space-guid", "process", "/")
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get application warning", "update warnings"))

					Expect(returnedApp).To(Equal(Application{
						GUID:            "some-app-guid",
						HealthCheckType: "process",
					}))

					Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
					app := fakeCloudControllerClient.UpdateApplicationArgsForCall(0)
					Expect(app).To(Equal(ccv2.Application{
						GUID:            "some-app-guid",
						HealthCheckType: "process",
					}))
				})
			})

			When("the desired health check type is 'http'", func() {
				When("the desired http endpoint is already set", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv2.Application{
								{GUID: "some-app-guid", HealthCheckType: "http", HealthCheckHTTPEndpoint: "/"},
							},
							ccv2.Warnings{"get application warning"},
							nil,
						)
					})

					It("does not send the update", func() {
						_, warnings, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
							"some-app", "some-space-guid", "http", "/")
						Expect(err).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("get application warning"))

						Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(0))
					})
				})

				When("the desired http endpoint is not set", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv2.Application{
								{GUID: "some-app-guid", HealthCheckType: "http", HealthCheckHTTPEndpoint: "/"},
							},
							ccv2.Warnings{"get application warning"},
							nil,
						)
						fakeCloudControllerClient.UpdateApplicationReturns(
							ccv2.Application{},
							ccv2.Warnings{"update warnings"},
							nil,
						)
					})

					It("sets the desired health check type and returns the warnings", func() {
						_, warnings, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
							"some-app", "some-space-guid", "http", "/v2/anything")
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
						app := fakeCloudControllerClient.UpdateApplicationArgsForCall(0)
						Expect(app).To(Equal(ccv2.Application{
							GUID:                    "some-app-guid",
							HealthCheckType:         "http",
							HealthCheckHTTPEndpoint: "/v2/anything",
						}))

						Expect(warnings).To(ConsistOf("get application warning", "update warnings"))
					})
				})
			})

			When("the application health check type is already set to the desired type", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{
							{
								GUID:            "some-app-guid",
								HealthCheckType: "process",
							},
						},
						ccv2.Warnings{"get application warning"},
						nil,
					)
				})

				It("does not update the health check type", func() {
					returnedApp, warnings, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
						"some-app", "some-space-guid", "process", "/")
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get application warning"))
					Expect(returnedApp).To(Equal(Application{
						GUID:            "some-app-guid",
						HealthCheckType: "process",
					}))

					Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(0))
				})
			})
		})

		When("getting the application returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{}, ccv2.Warnings{"get application warning"}, errors.New("get application error"))
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
					"some-app", "some-space-guid", "process", "/")

				Expect(warnings).To(ConsistOf("get application warning"))
				Expect(err).To(MatchError("get application error"))
			})
		})

		When("updating the application returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("foo bar")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{
						{GUID: "some-app-guid"},
					},
					ccv2.Warnings{"get application warning"},
					nil,
				)
				fakeCloudControllerClient.UpdateApplicationReturns(
					ccv2.Application{},
					ccv2.Warnings{"update warnings"},
					expectedErr,
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
					"some-app", "some-space-guid", "process", "/")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get application warning", "update warnings"))
			})
		})
	})

	Describe("StartApplication/RestartApplication", func() {
		var (
			app            Application
			fakeNOAAClient *v2actionfakes.FakeNOAAClient

			messages <-chan *LogMessage
			logErrs  <-chan error
			appState <-chan ApplicationStateChange
			warnings <-chan string
			errs     <-chan error

			eventStream chan *events.LogMessage
			errStream   chan error
		)

		BeforeEach(func() {
			fakeConfig.StagingTimeoutReturns(time.Minute)
			fakeConfig.StartupTimeoutReturns(time.Minute)

			app = Application{
				GUID:      "some-app-guid",
				Name:      "some-app",
				Instances: types.NullInt{Value: 2, IsSet: true},
			}

			fakeNOAAClient = new(v2actionfakes.FakeNOAAClient)
			fakeNOAAClient.TailingLogsStub = func(_ string, _ string) (<-chan *events.LogMessage, <-chan error) {
				eventStream = make(chan *events.LogMessage)
				errStream = make(chan error)
				return eventStream, errStream
			}

			closed := false
			fakeNOAAClient.CloseStub = func() error {
				if !closed {
					closed = true
					close(errStream)
					close(eventStream)
				}
				return nil
			}

			appCount := 0
			fakeCloudControllerClient.GetApplicationStub = func(appGUID string) (ccv2.Application, ccv2.Warnings, error) {
				if appCount == 0 {
					appCount++
					return ccv2.Application{
						GUID:         "some-app-guid",
						Instances:    types.NullInt{Value: 2, IsSet: true},
						Name:         "some-app",
						PackageState: constant.ApplicationPackagePending,
					}, ccv2.Warnings{"app-warnings-1"}, nil
				}

				return ccv2.Application{
					GUID:         "some-app-guid",
					Name:         "some-app",
					Instances:    types.NullInt{Value: 2, IsSet: true},
					PackageState: constant.ApplicationPackageStaged,
				}, ccv2.Warnings{"app-warnings-2"}, nil
			}

			instanceCount := 0
			fakeCloudControllerClient.GetApplicationApplicationInstancesStub = func(guid string) (map[int]ccv2.ApplicationInstance, ccv2.Warnings, error) {
				if instanceCount == 0 {
					instanceCount++
					return map[int]ccv2.ApplicationInstance{
						0: {State: constant.ApplicationInstanceStarting},
						1: {State: constant.ApplicationInstanceStarting},
					}, ccv2.Warnings{"app-instance-warnings-1"}, nil
				}

				return map[int]ccv2.ApplicationInstance{
					0: {State: constant.ApplicationInstanceStarting},
					1: {State: constant.ApplicationInstanceRunning},
				}, ccv2.Warnings{"app-instance-warnings-2"}, nil
			}
		})

		AfterEach(func() {
			Eventually(messages).Should(BeClosed())
			Eventually(logErrs).Should(BeClosed())
			Eventually(appState).Should(BeClosed())
			Eventually(warnings).Should(BeClosed())
			Eventually(errs).Should(BeClosed())
		})

		var ItHandlesStagingIssues = func() {
			Context("staging issues", func() {
				When("polling fails", func() {
					var expectedErr error
					BeforeEach(func() {
						expectedErr = errors.New("I am a banana!!!!")
						fakeCloudControllerClient.GetApplicationStub = func(appGUID string) (ccv2.Application, ccv2.Warnings, error) {
							return ccv2.Application{}, ccv2.Warnings{"app-warnings-1"}, expectedErr
						}
					})

					It("sends the error and stops polling", func() {
						Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
						Eventually(warnings).Should(Receive(Equal("state-warning")))
						Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
						Eventually(errs).Should(Receive(MatchError(expectedErr)))

						Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(0))
					})
				})

				When("the application fails to stage", func() {
					Context("due to a NoAppDetectedError", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationStub = func(appGUID string) (ccv2.Application, ccv2.Warnings, error) {
								return ccv2.Application{
									GUID:                "some-app-guid",
									Name:                "some-app",
									Instances:           types.NullInt{Value: 2, IsSet: true},
									PackageState:        constant.ApplicationPackageFailed,
									StagingFailedReason: "NoAppDetectedError",
								}, ccv2.Warnings{"app-warnings-1"}, nil
							}
						})

						It("sends a StagingFailedNoAppDetectedError and stops polling", func() {
							Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
							Eventually(warnings).Should(Receive(Equal("state-warning")))
							Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
							Eventually(errs).Should(Receive(MatchError(actionerror.StagingFailedNoAppDetectedError{Reason: "NoAppDetectedError"})))

							Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(0))
							Expect(fakeConfig.StagingTimeoutCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(0))
						})
					})

					Context("due to any other error", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationStub = func(appGUID string) (ccv2.Application, ccv2.Warnings, error) {
								return ccv2.Application{
									GUID:                "some-app-guid",
									Name:                "some-app",
									Instances:           types.NullInt{Value: 2, IsSet: true},
									PackageState:        constant.ApplicationPackageFailed,
									StagingFailedReason: "OhNoes",
								}, ccv2.Warnings{"app-warnings-1"}, nil
							}
						})

						It("sends a StagingFailedError and stops polling", func() {
							Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
							Eventually(warnings).Should(Receive(Equal("state-warning")))
							Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
							Eventually(errs).Should(Receive(MatchError(actionerror.StagingFailedError{Reason: "OhNoes"})))

							Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(0))
							Expect(fakeConfig.StagingTimeoutCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(0))
						})
					})
				})

				When("the application takes too long to stage", func() {
					BeforeEach(func() {
						fakeConfig.StagingTimeoutReturns(0)
						fakeCloudControllerClient.GetApplicationApplicationInstancesStub = nil
					})

					It("sends a timeout error and stops polling", func() {
						Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
						Eventually(warnings).Should(Receive(Equal("state-warning")))
						Eventually(errs).Should(Receive(MatchError(actionerror.StagingTimeoutError{AppName: "some-app", Timeout: 0})))

						Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(0))
						Expect(fakeConfig.StagingTimeoutCallCount()).To(Equal(2))
						Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(0))
					})
				})
			})
		}

		var ItHandlesStartingIssues = func() {
			Context("starting issues", func() {
				When("polling fails", func() {
					var expectedErr error
					BeforeEach(func() {
						expectedErr = errors.New("I am a banana!!!!")
						fakeCloudControllerClient.GetApplicationApplicationInstancesStub = func(guid string) (map[int]ccv2.ApplicationInstance, ccv2.Warnings, error) {
							return nil, ccv2.Warnings{"app-instance-warnings-1"}, expectedErr
						}
					})

					It("sends the error and stops polling", func() {
						Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
						Eventually(warnings).Should(Receive(Equal("state-warning")))
						Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
						Eventually(warnings).Should(Receive(Equal("app-warnings-2")))
						Eventually(appState).Should(Receive(Equal(ApplicationStateStarting)))
						Eventually(warnings).Should(Receive(Equal("app-instance-warnings-1")))
						Eventually(errs).Should(Receive(MatchError(expectedErr)))

						Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(1))
					})
				})

				When("the application takes too long to start", func() {
					BeforeEach(func() {
						fakeConfig.StartupTimeoutReturns(0)
					})

					It("sends a timeout error and stops polling", func() {
						Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
						Eventually(warnings).Should(Receive(Equal("state-warning")))
						Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
						Eventually(warnings).Should(Receive(Equal("app-warnings-2")))
						Eventually(appState).Should(Receive(Equal(ApplicationStateStarting)))
						Eventually(errs).Should(Receive(MatchError(actionerror.StartupTimeoutError{Name: "some-app"})))

						Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
						Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(0))
					})
				})

				When("the application crashes", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationApplicationInstancesStub = func(guid string) (map[int]ccv2.ApplicationInstance, ccv2.Warnings, error) {
							return map[int]ccv2.ApplicationInstance{
								0: {State: constant.ApplicationInstanceCrashed},
							}, ccv2.Warnings{"app-instance-warnings-1"}, nil
						}
					})

					It("returns an ApplicationInstanceCrashedError and stops polling", func() {
						Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
						Eventually(warnings).Should(Receive(Equal("state-warning")))
						Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
						Eventually(warnings).Should(Receive(Equal("app-warnings-2")))
						Eventually(appState).Should(Receive(Equal(ApplicationStateStarting)))
						Eventually(warnings).Should(Receive(Equal("app-instance-warnings-1")))
						Eventually(errs).Should(Receive(MatchError(actionerror.ApplicationInstanceCrashedError{Name: "some-app"})))

						Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
						Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(1))
					})
				})

				When("the application flaps", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationApplicationInstancesStub = func(guid string) (map[int]ccv2.ApplicationInstance, ccv2.Warnings, error) {
							return map[int]ccv2.ApplicationInstance{
								0: {State: constant.ApplicationInstanceFlapping},
							}, ccv2.Warnings{"app-instance-warnings-1"}, nil
						}
					})

					It("returns an ApplicationInstanceFlappingError and stops polling", func() {
						Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
						Eventually(warnings).Should(Receive(Equal("state-warning")))
						Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
						Eventually(warnings).Should(Receive(Equal("app-warnings-2")))
						Eventually(appState).Should(Receive(Equal(ApplicationStateStarting)))
						Eventually(warnings).Should(Receive(Equal("app-instance-warnings-1")))
						Eventually(errs).Should(Receive(MatchError(actionerror.ApplicationInstanceFlappingError{Name: "some-app"})))

						Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
						Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(1))
					})
				})
			})
		}

		var ItStartsApplication = func() {
			When("the app is not running", func() {
				It("starts and polls for an app instance", func() {
					Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
					Eventually(warnings).Should(Receive(Equal("state-warning")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-2")))
					Eventually(appState).Should(Receive(Equal(ApplicationStateStarting)))
					Eventually(warnings).Should(Receive(Equal("app-instance-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-instance-warnings-2")))

					Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(2))

					Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
					passedApp := fakeCloudControllerClient.UpdateApplicationArgsForCall(0)
					Expect(passedApp).To(Equal(ccv2.Application{
						GUID:  "some-app-guid",
						State: constant.ApplicationStarted,
					}))

					Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(2))
					Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(2))
					Eventually(fakeNOAAClient.CloseCallCount).Should(Equal(2))
				})
			})

			When("the app has zero instances", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateApplicationReturns(ccv2.Application{GUID: "some-app-guid",
						Instances: types.NullInt{Value: 0, IsSet: true},
						Name:      "some-app",
					}, ccv2.Warnings{"state-warning"}, nil)
				})

				It("starts and only polls for staging to finish", func() {
					Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
					Eventually(warnings).Should(Receive(Equal("state-warning")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-2")))
					Consistently(appState).ShouldNot(Receive(Equal(ApplicationStateStarting)))

					Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))

					Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
					passedApp := fakeCloudControllerClient.UpdateApplicationArgsForCall(0)
					Expect(passedApp).To(Equal(ccv2.Application{
						GUID:  "some-app-guid",
						State: constant.ApplicationStarted,
					}))

					Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(2))
					Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(0))
				})
			})

			When("updating the application fails", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("I am a banana!!!!")
					fakeCloudControllerClient.UpdateApplicationReturns(ccv2.Application{}, ccv2.Warnings{"state-warning"}, expectedErr)
				})

				It("sends the update error and never polls", func() {
					Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
					Eventually(warnings).Should(Receive(Equal("state-warning")))
					Eventually(errs).Should(Receive(MatchError(expectedErr)))

					Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(0))
				})
			})

			ItHandlesStagingIssues()

			ItHandlesStartingIssues()
		}

		Describe("StartApplication", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateApplicationReturns(ccv2.Application{GUID: "some-app-guid",
					Instances: types.NullInt{Value: 2, IsSet: true},
					Name:      "some-app",
				}, ccv2.Warnings{"state-warning"}, nil)
			})

			JustBeforeEach(func() {
				messages, logErrs, appState, warnings, errs = actor.StartApplication(app, fakeNOAAClient)
			})

			When("the app is already staged", func() {
				BeforeEach(func() {
					app.PackageState = constant.ApplicationPackageStaged
				})

				It("does not send ApplicationStateStaging", func() {
					Consistently(appState).ShouldNot(Receive(Equal(ApplicationStateStaging)))
					Eventually(warnings).Should(Receive(Equal("state-warning")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-2")))
					Eventually(appState).Should(Receive(Equal(ApplicationStateStarting)))
					Eventually(warnings).Should(Receive(Equal("app-instance-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-instance-warnings-2")))

					Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
					passedApp := fakeCloudControllerClient.UpdateApplicationArgsForCall(0)
					Expect(passedApp).To(Equal(ccv2.Application{
						GUID:  "some-app-guid",
						State: constant.ApplicationStarted,
					}))
				})
			})

			ItStartsApplication()
		})

		Describe("RestartApplication", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateApplicationReturns(ccv2.Application{GUID: "some-app-guid",
					Instances: types.NullInt{Value: 2, IsSet: true},
					Name:      "some-app",
				}, ccv2.Warnings{"state-warning"}, nil)
			})

			JustBeforeEach(func() {
				messages, logErrs, appState, warnings, errs = actor.RestartApplication(app, fakeNOAAClient)
			})

			When("application is running", func() {
				BeforeEach(func() {
					app.State = constant.ApplicationStarted
				})

				It("stops, starts and polls for an app instance", func() {
					Eventually(appState).Should(Receive(Equal(ApplicationStateStopping)))
					Eventually(warnings).Should(Receive(Equal("state-warning")))
					Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
					Eventually(warnings).Should(Receive(Equal("state-warning")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-2")))
					Eventually(appState).Should(Receive(Equal(ApplicationStateStarting)))
					Eventually(warnings).Should(Receive(Equal("app-instance-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-instance-warnings-2")))

					Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(2))

					Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(2))
					passedApp := fakeCloudControllerClient.UpdateApplicationArgsForCall(0)
					Expect(passedApp).To(Equal(ccv2.Application{
						GUID:  "some-app-guid",
						State: constant.ApplicationStopped,
					}))

					passedApp = fakeCloudControllerClient.UpdateApplicationArgsForCall(1)
					Expect(passedApp).To(Equal(ccv2.Application{
						GUID:  "some-app-guid",
						State: constant.ApplicationStarted,
					}))

					Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(2))
					Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(2))
					Eventually(fakeNOAAClient.CloseCallCount).Should(Equal(2))
				})

				When("updating the application to stop fails", func() {
					var expectedErr error
					BeforeEach(func() {
						expectedErr = errors.New("I am a banana!!!!")
						updateApplicationCalled := false
						fakeCloudControllerClient.UpdateApplicationStub = func(app ccv2.Application) (ccv2.Application, ccv2.Warnings, error) {
							if !updateApplicationCalled {
								return ccv2.Application{}, ccv2.Warnings{"state-warning"}, expectedErr
							}

							updateApplicationCalled = true
							return ccv2.Application{GUID: "some-app-guid",
								Instances: types.NullInt{Value: 2, IsSet: true},
								Name:      "some-app",
							}, ccv2.Warnings{"state-warning"}, nil
						}
					})

					It("sends the update error and never polls", func() {
						Eventually(appState).Should(Receive(Equal(ApplicationStateStopping)))
						Eventually(warnings).Should(Receive(Equal("state-warning")))
						Eventually(errs).Should(Receive(MatchError(expectedErr)))
						Eventually(appState).ShouldNot(Receive(Equal(ApplicationStateStaging)))

						Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(0))
					})
				})
			})

			When("the app is not running", func() {
				BeforeEach(func() {
					app.State = constant.ApplicationStopped
				})

				It("does not stop an app instance", func() {
					Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
					Eventually(warnings).Should(Receive(Equal("state-warning")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-2")))
					Eventually(appState).Should(Receive(Equal(ApplicationStateStarting)))
					Eventually(warnings).Should(Receive(Equal("app-instance-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-instance-warnings-2")))

					Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
					passedApp := fakeCloudControllerClient.UpdateApplicationArgsForCall(0)
					Expect(passedApp).To(Equal(ccv2.Application{
						GUID:  "some-app-guid",
						State: constant.ApplicationStarted,
					}))
				})
			})

			When("the app is already staged", func() {
				BeforeEach(func() {
					app.PackageState = constant.ApplicationPackageStaged
				})

				It("does not send ApplicationStateStaging", func() {
					Consistently(appState).ShouldNot(Receive(Equal(ApplicationStateStaging)))
					Eventually(warnings).Should(Receive(Equal("state-warning")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-2")))
					Eventually(appState).Should(Receive(Equal(ApplicationStateStarting)))
					Eventually(warnings).Should(Receive(Equal("app-instance-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-instance-warnings-2")))

					Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
					passedApp := fakeCloudControllerClient.UpdateApplicationArgsForCall(0)
					Expect(passedApp).To(Equal(ccv2.Application{
						GUID:  "some-app-guid",
						State: constant.ApplicationStarted,
					}))
				})
			})

			ItStartsApplication()
		})

		Describe("RestageApplication", func() {
			JustBeforeEach(func() {
				messages, logErrs, appState, warnings, errs = actor.RestageApplication(app, fakeNOAAClient)
			})

			When("restaging succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.RestageApplicationReturns(ccv2.Application{GUID: "some-app-guid",
						Instances: types.NullInt{Value: 2, IsSet: true},
						Name:      "some-app",
					}, ccv2.Warnings{"state-warning"}, nil)
				})

				It("restages and polls for app instances", func() {
					Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
					Eventually(warnings).Should(Receive(Equal("state-warning")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-warnings-2")))
					Eventually(appState).Should(Receive(Equal(ApplicationStateStarting)))
					Eventually(warnings).Should(Receive(Equal("app-instance-warnings-1")))
					Eventually(warnings).Should(Receive(Equal("app-instance-warnings-2")))

					Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(2))

					Expect(fakeCloudControllerClient.RestageApplicationCallCount()).To(Equal(1))
					app := fakeCloudControllerClient.RestageApplicationArgsForCall(0)
					Expect(app).To(Equal(ccv2.Application{
						GUID: "some-app-guid",
					}))

					Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(2))
					Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(2))
					Eventually(fakeNOAAClient.CloseCallCount).Should(Equal(2))
				})

				ItHandlesStagingIssues()

				ItHandlesStartingIssues()
			})

			When("restaging errors", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.RestageApplicationReturns(ccv2.Application{GUID: "some-app-guid",
						Instances: types.NullInt{Value: 2, IsSet: true},
						Name:      "some-app",
					}, ccv2.Warnings{"state-warning"}, errors.New("some-error"))
				})

				It("sends the restage error and never polls", func() {
					Eventually(appState).Should(Receive(Equal(ApplicationStateStaging)))
					Eventually(warnings).Should(Receive(Equal("state-warning")))
					Eventually(errs).Should(Receive(MatchError("some-error")))

					Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(0))
				})
			})
		})
	})

	Describe("UpdateApplication", func() {
		When("the update is successful", func() {
			var expectedApp ccv2.Application
			BeforeEach(func() {
				expectedApp = ccv2.Application{
					GUID:      "some-app-guid",
					Name:      "some-app-name",
					SpaceGUID: "some-space-guid",
				}
				fakeCloudControllerClient.UpdateApplicationReturns(expectedApp, ccv2.Warnings{"some-app-warning-1"}, nil)
			})

			It("updates and returns the application", func() {
				newApp := Application{
					Name:      "some-app-name",
					SpaceGUID: "some-space-guid",
				}
				app, warnings, err := actor.UpdateApplication(newApp)
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-app-warning-1"))
				Expect(app).To(Equal(Application(expectedApp)))

				Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateApplicationArgsForCall(0)).To(Equal(ccv2.Application(newApp)))
			})
		})

		When("the client returns back an error", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some update app error")
				fakeCloudControllerClient.UpdateApplicationReturns(ccv2.Application{}, ccv2.Warnings{"some-app-warning-1"}, expectedErr)
			})

			It("returns warnings and an error", func() {
				newApp := Application{
					Name:      "some-app-name",
					SpaceGUID: "some-space-guid",
				}
				_, warnings, err := actor.UpdateApplication(newApp)
				Expect(warnings).To(ConsistOf("some-app-warning-1"))
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})
})
