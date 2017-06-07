package v3action_test

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
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
		actor = NewActor(fakeCloudControllerClient, fakeConfig)
	})

	Describe("GetApplicationByNameAndSpace", func() {
		Context("when the app exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name: "some-app-name",
							GUID: "some-app-guid",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the application and warnings", func() {
				app, warnings, err := actor.GetApplicationByNameAndSpace("some-app-name", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(app).To(Equal(Application{
					Name: "some-app-name",
					GUID: "some-app-guid",
				}))
				Expect(warnings).To(Equal(Warnings{"some-warning"}))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				expectedQuery := url.Values{
					"names":       []string{"some-app-name"},
					"space_guids": []string{"some-space-guid"},
				}
				query := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
				Expect(query).To(Equal(expectedQuery))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				_, warnings, err := actor.GetApplicationByNameAndSpace("some-app-name", "some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				expectedQuery := url.Values{
					"names":       []string{"some-app-name"},
					"space_guids": []string{"some-space-guid"},
				}
				query := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
				Expect(query).To(Equal(expectedQuery))
			})
		})

		Context("when the app does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns an ApplicationNotFoundError and the warnings", func() {
				_, warnings, err := actor.GetApplicationByNameAndSpace("some-app-name", "some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(
					ApplicationNotFoundError{Name: "some-app-name"}))
				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				expectedQuery := url.Values{
					"names":       []string{"some-app-name"},
					"space_guids": []string{"some-space-guid"},
				}
				query := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
				Expect(query).To(Equal(expectedQuery))
			})
		})
	})

	Describe("CreateApplicationByNameAndSpace", func() {
		Context("when the app successfully gets created", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationReturns(
					ccv3.Application{
						Name: "some-app-name",
						GUID: "some-app-guid",
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("creates and returns the application and warnings", func() {
				app, warnings, err := actor.CreateApplicationByNameAndSpace("some-app-name", "some-space-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(app).To(Equal(Application{
					Name: "some-app-name",
					GUID: "some-app-guid",
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.CreateApplicationCallCount()).To(Equal(1))
				expectedApp := ccv3.Application{
					Name: "some-app-name",
					Relationships: ccv3.Relationships{
						ccv3.SpaceRelationship: ccv3.Relationship{GUID: "some-space-guid"},
					},
				}
				Expect(fakeCloudControllerClient.CreateApplicationArgsForCall(0)).To(Equal(expectedApp))
			})
		})

		Context("when the cc client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.CreateApplicationReturns(
					ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError,
				)
			})

			It("raises the error and warnings", func() {
				_, warnings, err := actor.CreateApplicationByNameAndSpace("some-app-name", "some-space-guid")

				Expect(err).To(MatchError(expectedError))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})

		Context("when the cc client response contains an UnprocessableEntityError", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationReturns(
					ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					ccerror.UnprocessableEntityError{},
				)
			})

			It("raises the error as ApplicationAlreadyExistsError and warnings", func() {
				_, warnings, err := actor.CreateApplicationByNameAndSpace("some-app-name", "some-space-guid")

				Expect(err).To(MatchError(ApplicationAlreadyExistsError{Name: "some-app-name"}))
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

		Context("when getting the application processes fails", func() {
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

		Context("when getting the application processes succeeds", func() {
			var processes []ccv3.Process

			BeforeEach(func() {
				fakeConfig.StartupTimeoutReturns(time.Second)
				fakeConfig.PollingIntervalReturns(0)
			})

			JustBeforeEach(func() {
				fakeCloudControllerClient.GetApplicationProcessesReturns(
					processes,
					ccv3.Warnings{"get-app-warning-1"}, nil)
			})

			Context("when there is a single process", func() {
				BeforeEach(func() {
					processes = []ccv3.Process{{GUID: "abc123"}}
				})

				Context("when the polling times out", func() {
					BeforeEach(func() {
						fakeConfig.StartupTimeoutReturns(time.Millisecond)
						fakeConfig.PollingIntervalReturns(time.Millisecond * 2)
						fakeCloudControllerClient.GetProcessInstancesReturns(
							[]ccv3.Instance{{State: "STARTING"}},
							ccv3.Warnings{"get-process-warning-1", "get-process-warning-2"},
							nil,
						)
					})

					It("returns the timeout error", func() {
						err := actor.PollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2"))
						Expect(err).To(MatchError(StartupTimeoutError{}))
					})

					It("gets polling and timeout values from the config", func() {
						actor.PollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
						Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
					})
				})

				Context("when getting the process instances errors", func() {
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

				Context("when getting the process instances succeeds", func() {
					var (
						initialInstanceStates    []ccv3.Instance
						eventualInstanceStates   []ccv3.Instance
						pollStartErr             error
						processInstanceCallCount int
					)

					BeforeEach(func() {
						processInstanceCallCount = 0
					})

					JustBeforeEach(func() {
						fakeCloudControllerClient.GetProcessInstancesStub = func(processGuid string) ([]ccv3.Instance, ccv3.Warnings, error) {
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

					Context("when there are no process instances", func() {
						BeforeEach(func() {
							initialInstanceStates = []ccv3.Instance{}
							eventualInstanceStates = []ccv3.Instance{}
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

					Context("when all instances become running by the second call", func() {
						BeforeEach(func() {
							initialInstanceStates = []ccv3.Instance{{State: "STARTING"}, {State: "STARTING"}}
							eventualInstanceStates = []ccv3.Instance{{State: "RUNNING"}, {State: "RUNNING"}}
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

					Context("when at least one instance has become running by the second call", func() {
						BeforeEach(func() {
							initialInstanceStates = []ccv3.Instance{{State: "STARTING"}, {State: "STARTING"}, {State: "STARTING"}}
							eventualInstanceStates = []ccv3.Instance{{State: "STARTING"}, {State: "STARTING"}, {State: "RUNNING"}}
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
					fakeCloudControllerClient.GetProcessInstancesStub = func(processGuid string) ([]ccv3.Instance, ccv3.Warnings, error) {
						defer func() { processInstanceCallCount++ }()
						if strings.HasPrefix(processGuid, "good") {
							return []ccv3.Instance{ccv3.Instance{State: "RUNNING"}}, nil, nil
						} else {
							return []ccv3.Instance{ccv3.Instance{State: "STARTING"}}, nil, nil
						}
					}

					pollStartErr = actor.PollStart("some-guid", warningsChannel)
					funcDone <- nil
				})

				Context("when none of the processes are ready", func() {
					BeforeEach(func() {
						processes = []ccv3.Process{{GUID: "bad-1"}, {GUID: "bad-2"}}
					})

					It("returns the timeout error", func() {
						Expect(pollStartErr).To(MatchError(StartupTimeoutError{}))
					})

				})

				Context("when some of the processes are ready", func() {
					BeforeEach(func() {
						processes = []ccv3.Process{{GUID: "bad-1"}, {GUID: "good-1"}}
					})

					It("returns the timeout error", func() {
						Expect(pollStartErr).To(MatchError(StartupTimeoutError{}))
					})
				})

				Context("when all of the processes are ready", func() {
					BeforeEach(func() {
						processes = []ccv3.Process{{GUID: "good-1"}, {GUID: "good-2"}}
					})

					It("returns nil", func() {
						Expect(pollStartErr).ToNot(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("StartApplication", func() {
		Context("when there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.StartApplicationReturns(
					ccv3.Application{GUID: "some-app-guid"},
					ccv3.Warnings{"start-application-warning"},
					nil,
				)
			})

			It("sets the app's droplet", func() {
				app, warnings, err := actor.StartApplication("some-app-name", "some-space-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-applications-warning", "start-application-warning"))
				Expect(app).To(Equal(Application{GUID: "some-app-guid"}))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				queryURL := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
				query := url.Values{"names": []string{"some-app-name"}, "space_guids": []string{"some-space-guid"}}
				Expect(queryURL).To(Equal(query))

				Expect(fakeCloudControllerClient.StartApplicationCallCount()).To(Equal(1))
				appGUID := fakeCloudControllerClient.StartApplicationArgsForCall(0)
				Expect(appGUID).To(Equal("some-app-guid"))
			})
		})

		Context("when getting the application fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some get application error")

				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"get-applications-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.StartApplication("some-app-name", "some-space-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-applications-warning"))
			})
		})

		Context("when starting the application fails", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some set start-application error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.StartApplicationReturns(
					ccv3.Application{},
					ccv3.Warnings{"start-application-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.StartApplication("some-app-name", "some-space-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-applications-warning", "start-application-warning"))
			})
		})
	})
})
