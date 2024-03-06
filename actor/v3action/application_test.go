package v3action_test

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
		fakeConfig                *v3actionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v3actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil)
	})

	Describe("DeleteApplicationByNameAndSpace", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.DeleteApplicationByNameAndSpace("some-app", "some-space-guid")
		})

		When("looking up the app guid fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{}, ccv3.Warnings{"some-get-app-warning"}, errors.New("some-get-app-error"))
			})

			It("returns the warnings and error", func() {
				Expect(warnings).To(ConsistOf("some-get-app-warning"))
				Expect(executeErr).To(MatchError("some-get-app-error"))
			})
		})

		When("looking up the app guid succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{{Name: "some-app", GUID: "abc123"}}, ccv3.Warnings{"some-get-app-warning"}, nil)
			})

			When("sending the delete fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteApplicationReturns("", ccv3.Warnings{"some-delete-app-warning"}, errors.New("some-delete-app-error"))
				})

				It("returns the warnings and error", func() {
					Expect(warnings).To(ConsistOf("some-get-app-warning", "some-delete-app-warning"))
					Expect(executeErr).To(MatchError("some-delete-app-error"))
				})
			})

			When("sending the delete succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteApplicationReturns("/some-job-url", ccv3.Warnings{"some-delete-app-warning"}, nil)
				})

				When("polling fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-poll-warning"}, errors.New("some-poll-error"))
					})

					It("returns the warnings and poll error", func() {
						Expect(warnings).To(ConsistOf("some-get-app-warning", "some-delete-app-warning", "some-poll-warning"))
						Expect(executeErr).To(MatchError("some-poll-error"))
					})
				})

				When("polling succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-poll-warning"}, nil)
					})

					It("returns all the warnings and no error", func() {
						Expect(warnings).To(ConsistOf("some-get-app-warning", "some-delete-app-warning", "some-poll-warning"))
						Expect(executeErr).ToNot(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("GetApplicationByNameAndSpace", func() {
		When("the app exists", func() {
			var (
				warnings   Warnings
				executeErr error
				app        Application
			)

			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							Name:      "some-app-name",
							GUID:      "some-app-guid",
							SpaceGUID: "some-space-guid",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			JustBeforeEach(func() {
				app, warnings, executeErr = actor.GetApplicationByNameAndSpace("some-app-name", "some-space-guid")
			})

			It("returns the application and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(app).To(Equal(Application{
					Name:      "some-app-name",
					GUID:      "some-app-guid",
					SpaceGUID: "some-space-guid",
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name"}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
				))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				_, warnings, err := actor.GetApplicationByNameAndSpace("some-app-name", "some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
			})
		})

		When("the app does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns an ApplicationNotFoundError and the warnings", func() {
				_, warnings, err := actor.GetApplicationByNameAndSpace("some-app-name", "some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(actionerror.ApplicationNotFoundError{Name: "some-app-name"}))
			})
		})
	})

	Describe("GetApplicationsBySpace", func() {
		When("there are applications in the space", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							GUID:      "some-app-guid-1",
							Name:      "some-app-1",
							SpaceGUID: "some-space-guid-1",
						},
						{
							GUID:      "some-app-guid-2",
							Name:      "some-app-2",
							SpaceGUID: "some-space-guid-2",
						},
					},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns the application and warnings", func() {
				apps, warnings, err := actor.GetApplicationsBySpace("some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(apps).To(ConsistOf(
					Application{
						GUID:      "some-app-guid-1",
						Name:      "some-app-1",
						SpaceGUID: "some-space-guid-1",
					},
					Application{
						GUID:      "some-app-guid-2",
						Name:      "some-app-2",
						SpaceGUID: "some-space-guid-2",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
				))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetApplicationsBySpace("some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("GetApplicationsByGUIDs", func() {
		When("there are applications that match provided guids", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							GUID:      "some-app-guid-1",
							Name:      "some-app-1",
							SpaceGUID: "some-space-guid-1",
						},
						{
							GUID:      "some-app-guid-2",
							Name:      "some-app-2",
							SpaceGUID: "some-space-guid-2",
						},
					},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns the applications and warnings", func() {
				apps, warnings, err := actor.GetApplicationsByGUIDs("some-app-guid-1", "some-app-guid-2")
				Expect(err).ToNot(HaveOccurred())
				Expect(apps).To(ConsistOf(
					Application{
						GUID:      "some-app-guid-1",
						Name:      "some-app-1",
						SpaceGUID: "some-space-guid-1",
					},
					Application{
						GUID:      "some-app-guid-2",
						Name:      "some-app-2",
						SpaceGUID: "some-space-guid-2",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{"some-app-guid-1", "some-app-guid-2"}},
				))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetApplicationsByGUIDs("some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("CreateApplicationInSpace", func() {
		var (
			application Application
			warnings    Warnings
			err         error
		)

		JustBeforeEach(func() {
			application, warnings, err = actor.CreateApplicationInSpace(Application{
				Name:                "some-app-name",
				LifecycleType:       constant.AppLifecycleTypeBuildpack,
				LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
			}, "some-space-guid")
		})

		When("the app successfully gets created", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationReturns(
					resources.Application{
						Name:                "some-app-name",
						GUID:                "some-app-guid",
						LifecycleType:       constant.AppLifecycleTypeBuildpack,
						LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
						SpaceGUID:           "some-space-guid",
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("creates and returns the application and warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(application).To(Equal(Application{
					Name:                "some-app-name",
					GUID:                "some-app-guid",
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
					SpaceGUID:           "some-space-guid",
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.CreateApplicationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateApplicationArgsForCall(0)).To(Equal(resources.Application{
					Name:                "some-app-name",
					SpaceGUID:           "some-space-guid",
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
				}))
			})
		})

		When("the cc client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.CreateApplicationReturns(
					resources.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError,
				)
			})

			It("raises the error and warnings", func() {
				Expect(err).To(MatchError(expectedError))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})

		When("the cc client returns an NameNotUniqueInSpaceError", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationReturns(
					resources.Application{},
					ccv3.Warnings{"some-warning"},
					ccerror.NameNotUniqueInSpaceError{},
				)
			})

			It("returns the NameNotUniqueInSpaceError and warnings", func() {
				Expect(err).To(MatchError(ccerror.NameNotUniqueInSpaceError{}))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})
	})

	Describe("UpdateApplication", func() {
		var (
			submitApp, resultApp Application
			warnings             Warnings
			err                  error
		)

		JustBeforeEach(func() {
			submitApp = Application{
				GUID:                "some-app-guid",
				StackName:           "some-stack-name",
				LifecycleType:       constant.AppLifecycleTypeBuildpack,
				LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
			}
			resultApp, warnings, err = actor.UpdateApplication(submitApp)
		})

		When("the app successfully gets updated", func() {
			var apiResponseApp resources.Application

			BeforeEach(func() {
				apiResponseApp = resources.Application{
					GUID:                "response-app-guid",
					StackName:           "response-stack-name",
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"response-buildpack-1", "response-buildpack-2"},
					SpaceGUID:           "some-space-guid",
				}
				fakeCloudControllerClient.UpdateApplicationReturns(
					apiResponseApp,
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("creates and returns the application and warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(resultApp).To(Equal(Application{
					GUID:                apiResponseApp.GUID,
					StackName:           apiResponseApp.StackName,
					LifecycleType:       apiResponseApp.LifecycleType,
					LifecycleBuildpacks: apiResponseApp.LifecycleBuildpacks,
					SpaceGUID:           "some-space-guid",
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateApplicationArgsForCall(0)).To(Equal(resources.Application{
					GUID:                submitApp.GUID,
					StackName:           submitApp.StackName,
					LifecycleType:       submitApp.LifecycleType,
					LifecycleBuildpacks: submitApp.LifecycleBuildpacks,
				}))
			})
		})

		When("the cc client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.UpdateApplicationReturns(
					resources.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError,
				)
			})

			It("raises the error and warnings", func() {
				Expect(err).To(MatchError(expectedError))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})
	})

	Describe("PollStart", func() {
		var warningsChannel chan Warnings
		var allWarnings Warnings
		var funcDone chan interface{}

		BeforeEach(func() {
			warningsChannel = make(chan Warnings)
			funcDone = make(chan interface{})
			allWarnings = Warnings{}
			go func() {
				for {
					select {
					case warnings := <-warningsChannel:
						allWarnings = append(allWarnings, warnings...)
					case <-funcDone:
						return
					}
				}
			}()
		})

		When("getting the application processes fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationProcessesReturns(nil, ccv3.Warnings{"get-app-warning-1", "get-app-warning-2"}, errors.New("some-error"))
			})

			It("returns the error and all warnings", func() {
				err := actor.PollStart("some-guid", warningsChannel)
				funcDone <- nil
				Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-app-warning-2"))
				Expect(err).To(MatchError(errors.New("some-error")))
			})
		})

		When("getting the application processes succeeds", func() {
			var processes []resources.Process

			BeforeEach(func() {
				fakeConfig.StartupTimeoutReturns(time.Second)
				fakeConfig.PollingIntervalReturns(0)
			})

			JustBeforeEach(func() {
				fakeCloudControllerClient.GetApplicationProcessesReturns(
					processes,
					ccv3.Warnings{"get-app-warning-1"}, nil)
			})

			When("there is a single process", func() {
				BeforeEach(func() {
					processes = []resources.Process{{GUID: "abc123"}}
				})

				When("the polling times out", func() {
					BeforeEach(func() {
						fakeConfig.StartupTimeoutReturns(time.Millisecond)
						fakeConfig.PollingIntervalReturns(time.Millisecond * 2)
						fakeCloudControllerClient.GetProcessInstancesReturns(
							[]ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}},
							ccv3.Warnings{"get-process-warning-1", "get-process-warning-2"},
							nil,
						)
					})

					It("returns the timeout error", func() {
						err := actor.PollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2"))
						Expect(err).To(MatchError(actionerror.StartupTimeoutError{}))
					})

					It("gets polling and timeout values from the config", func() {
						err := actor.PollStart("some-guid", warningsChannel)
						Expect(err).To(MatchError(actionerror.StartupTimeoutError{}))
						funcDone <- nil

						Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
						Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
					})
				})

				When("getting the process instances errors", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetProcessInstancesReturns(
							nil,
							ccv3.Warnings{"get-process-warning-1", "get-process-warning-2"},
							errors.New("some-error"),
						)
					})

					It("returns the error", func() {
						err := actor.PollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2"))
						Expect(err).To(MatchError("some-error"))
					})
				})

				When("getting the process instances succeeds", func() {
					var (
						initialInstanceStates    []ccv3.ProcessInstance
						eventualInstanceStates   []ccv3.ProcessInstance
						pollStartErr             error
						processInstanceCallCount int
					)

					BeforeEach(func() {
						processInstanceCallCount = 0
					})

					JustBeforeEach(func() {
						fakeCloudControllerClient.GetProcessInstancesStub = func(processGuid string) ([]ccv3.ProcessInstance, ccv3.Warnings, error) {
							defer func() { processInstanceCallCount++ }()
							if processInstanceCallCount == 0 {
								return initialInstanceStates,
									ccv3.Warnings{"get-process-warning-1", "get-process-warning-2"},
									nil
							} else {
								return eventualInstanceStates,
									ccv3.Warnings{fmt.Sprintf("get-process-warning-%d", processInstanceCallCount+2)},
									nil
							}
						}

						pollStartErr = actor.PollStart("some-guid", warningsChannel)
						funcDone <- nil
					})

					When("there are no process instances", func() {
						BeforeEach(func() {
							initialInstanceStates = []ccv3.ProcessInstance{}
							eventualInstanceStates = []ccv3.ProcessInstance{}
						})

						It("should not return an error", func() {
							Expect(pollStartErr).NotTo(HaveOccurred())
						})

						It("should only call GetProcessInstances once before exiting", func() {
							Expect(processInstanceCallCount).To(Equal(1))
						})

						It("should return correct warnings", func() {
							Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2"))
						})
					})

					When("all instances become running by the second call", func() {
						BeforeEach(func() {
							initialInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}}
							eventualInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceRunning}, {State: constant.ProcessInstanceRunning}}
						})

						It("should not return an error", func() {
							Expect(pollStartErr).NotTo(HaveOccurred())
						})

						It("should call GetProcessInstances twice", func() {
							Expect(processInstanceCallCount).To(Equal(2))
						})

						It("should return correct warnings", func() {
							Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2", "get-process-warning-3"))
						})
					})

					When("at least one instance has become running by the second call", func() {
						BeforeEach(func() {
							initialInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}}
							eventualInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceRunning}}
						})

						It("should not return an error", func() {
							Expect(pollStartErr).NotTo(HaveOccurred())
						})

						It("should call GetProcessInstances twice", func() {
							Expect(processInstanceCallCount).To(Equal(2))
						})

						It("should return correct warnings", func() {
							Expect(len(allWarnings)).To(Equal(4))
							Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2", "get-process-warning-3"))
						})
					})

					When("all of the instances have crashed by the second call", func() {
						BeforeEach(func() {
							initialInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}}
							eventualInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceCrashed}, {State: constant.ProcessInstanceCrashed}, {State: constant.ProcessInstanceCrashed}}
						})

						It("should not return an error", func() {
							Expect(pollStartErr).NotTo(HaveOccurred())
						})

						It("should call GetProcessInstances twice", func() {
							Expect(processInstanceCallCount).To(Equal(2))
						})

						It("should return correct warnings", func() {
							Expect(len(allWarnings)).To(Equal(4))
							Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2", "get-process-warning-3"))
						})
					})
				})
			})

			Context("where there are multiple processes", func() {
				var (
					pollStartErr             error
					processInstanceCallCount int
				)

				BeforeEach(func() {
					processInstanceCallCount = 0
					fakeConfig.StartupTimeoutReturns(time.Millisecond)
					fakeConfig.PollingIntervalReturns(time.Millisecond * 2)
				})

				JustBeforeEach(func() {
					fakeCloudControllerClient.GetProcessInstancesStub = func(processGuid string) ([]ccv3.ProcessInstance, ccv3.Warnings, error) {
						defer func() { processInstanceCallCount++ }()
						if strings.HasPrefix(processGuid, "good") {
							return []ccv3.ProcessInstance{{State: constant.ProcessInstanceRunning}}, nil, nil
						} else {
							return []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}}, nil, nil
						}
					}

					pollStartErr = actor.PollStart("some-guid", warningsChannel)
					funcDone <- nil
				})

				When("none of the processes are ready", func() {
					BeforeEach(func() {
						processes = []resources.Process{{GUID: "bad-1"}, {GUID: "bad-2"}}
					})

					It("returns the timeout error", func() {
						Expect(pollStartErr).To(MatchError(actionerror.StartupTimeoutError{}))
					})

				})

				When("some of the processes are ready", func() {
					BeforeEach(func() {
						processes = []resources.Process{{GUID: "bad-1"}, {GUID: "good-1"}}
					})

					It("returns the timeout error", func() {
						Expect(pollStartErr).To(MatchError(actionerror.StartupTimeoutError{}))
					})
				})

				When("all of the processes are ready", func() {
					BeforeEach(func() {
						processes = []resources.Process{{GUID: "good-1"}, {GUID: "good-2"}}
					})

					It("returns nil", func() {
						Expect(pollStartErr).ToNot(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("StopApplication", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.StopApplication("some-app-guid")
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateApplicationStopReturns(
					resources.Application{GUID: "some-app-guid"},
					ccv3.Warnings{"stop-application-warning"},
					nil,
				)
			})

			It("stops the application", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("stop-application-warning"))

				Expect(fakeCloudControllerClient.UpdateApplicationStopCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateApplicationStopArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		When("stopping the application fails", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some set stop-application error")
				fakeCloudControllerClient.UpdateApplicationStopReturns(
					resources.Application{},
					ccv3.Warnings{"stop-application-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("stop-application-warning"))
			})
		})
	})

	Describe("StartApplication", func() {
		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateApplicationStartReturns(
					resources.Application{GUID: "some-app-guid"},
					ccv3.Warnings{"start-application-warning"},
					nil,
				)
			})

			It("starts the application", func() {
				warnings, err := actor.StartApplication("some-app-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("start-application-warning"))

				Expect(fakeCloudControllerClient.UpdateApplicationStartCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateApplicationStartArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		When("starting the application fails", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some set start-application error")
				fakeCloudControllerClient.UpdateApplicationStartReturns(
					resources.Application{},
					ccv3.Warnings{"start-application-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				warnings, err := actor.StartApplication("some-app-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("start-application-warning"))
			})
		})
	})

	Describe("RestartApplication", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.RestartApplication("some-app-guid")
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateApplicationRestartReturns(
					resources.Application{GUID: "some-app-guid"},
					ccv3.Warnings{"restart-application-warning"},
					nil,
				)
			})

			It("stops the application", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("restart-application-warning"))

				Expect(fakeCloudControllerClient.UpdateApplicationRestartCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateApplicationRestartArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		When("restarting the application fails", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some set restart-application error")
				fakeCloudControllerClient.UpdateApplicationRestartReturns(
					resources.Application{},
					ccv3.Warnings{"restart-application-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("restart-application-warning"))
			})
		})
	})

})
