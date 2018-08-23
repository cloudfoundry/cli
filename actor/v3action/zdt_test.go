package v3action_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("v3-zdt-push", func() {

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

	Describe("ZeroDowntimePollStart", func() {
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
				err := actor.ZeroDowntimePollStart("some-guid", warningsChannel)
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
				processes = []ccv3.Process{
					{GUID: "web-guid", Type: "web"},
					{GUID: "web-ish-guid", Type: "web-deployment-efg456"},
				}
			})

			JustBeforeEach(func() {
				fakeCloudControllerClient.GetApplicationProcessesReturns(
					processes,
					ccv3.Warnings{"get-app-warning-1"}, nil)
			})

			Context("when the polling times out", func() {
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
					err := actor.ZeroDowntimePollStart("some-guid", warningsChannel)
					funcDone <- nil

					Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2"))
					Expect(err).To(MatchError(actionerror.StartupTimeoutError{}))
				})

				It("gets polling and timeout values from the config", func() {
					actor.ZeroDowntimePollStart("some-guid", warningsChannel)
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
					err := actor.ZeroDowntimePollStart("some-guid", warningsChannel)
					funcDone <- nil

					Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2"))
					Expect(err).To(MatchError("some-error"))
				})
			})

			Context("when getting the process instances succeeds", func() {
				var (
					processInstanceCallCount  int
					processInstancesCallGuids []string
					initialInstanceStates     []ccv3.ProcessInstance
					eventualInstanceStates    []ccv3.ProcessInstance
					pollStartErr              error
				)

				BeforeEach(func() {
					processInstanceCallCount = 0
					processInstancesCallGuids = []string{}

					fakeCloudControllerClient.GetProcessInstancesStub = func(processGuid string) ([]ccv3.ProcessInstance, ccv3.Warnings, error) {
						processInstancesCallGuids = append(processInstancesCallGuids, processGuid)
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
				})

				Context("when there are no instances for the deploying process", func() {
					BeforeEach(func() {
						initialInstanceStates = []ccv3.ProcessInstance{}
					})

					It("should not return an error", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(pollStartErr).NotTo(HaveOccurred())
					})

					It("should only call GetProcessInstances once before exiting", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(processInstanceCallCount).To(Equal(1))
					})

					It("should return correct warnings", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2"))
					})
				})

				Context("when the deploying process has at least one running instance by the second call", func() {
					BeforeEach(func() {
						initialInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}}
						eventualInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceRunning}}
					})

					It("should not return an error", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(pollStartErr).NotTo(HaveOccurred())
					})

					It("should call GetProcessInstances twice", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(processInstanceCallCount).To(Equal(2))
					})

					It("should return correct warnings", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2", "get-process-warning-3"))
					})

					It("should only call GetProcessInstances for the webish process", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(processInstancesCallGuids).To(ConsistOf("web-ish-guid", "web-ish-guid"))
					})
				})

				Context("when there is no webish process", func() {
					BeforeEach(func() {
						fakeConfig.StartupTimeoutReturns(time.Second)
						fakeConfig.PollingIntervalReturns(0)
						processes = []ccv3.Process{
							{GUID: "web-guid", Type: "web"},
							{GUID: "worker-guid", Type: "worker"},
						}
					})

					It("should not return an error", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(pollStartErr).NotTo(HaveOccurred())
					})

					It("should call not call GetProcessInstances, because the deploy has already succeeded", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(processInstanceCallCount).To(Equal(0))
					})

					It("should return correct warnings", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(allWarnings).To(ConsistOf("get-app-warning-1"))
					})
				})

				Context("when all of the instances have crashed by the second call", func() {
					BeforeEach(func() {
						initialInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}}
						eventualInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceCrashed}, {State: constant.ProcessInstanceCrashed}, {State: constant.ProcessInstanceCrashed}}
					})

					It("should not return an error", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(pollStartErr).NotTo(HaveOccurred())
					})

					It("should call GetProcessInstances twice", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(processInstanceCallCount).To(Equal(2))
					})

					It("should return correct warnings", func() {
						pollStartErr = actor.ZeroDowntimePollStart("some-guid", warningsChannel)
						funcDone <- nil

						Expect(allWarnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2", "get-process-warning-3"))
					})
				})
			})

		})
	})
})
