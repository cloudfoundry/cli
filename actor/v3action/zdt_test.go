package v3action_test

import (
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

	Describe("CancelDeploymentByAppNameAndSpace", func() {
		var (
			app resources.Application
		)

		BeforeEach(func() {
			app = resources.Application{GUID: "app-guid"}
			fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{app}, ccv3.Warnings{"getapp-warning"}, nil)
			fakeCloudControllerClient.GetDeploymentsReturns([]resources.Deployment{{GUID: "deployment-guid"}}, ccv3.Warnings{"getdep-warning"}, nil)
			fakeCloudControllerClient.CancelDeploymentReturns(ccv3.Warnings{"cancel-warning"}, nil)
		})

		It("cancels the appropriate deployment", func() {
			warnings, err := actor.CancelDeploymentByAppNameAndSpace("app-name", "space-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf(Warnings{"getapp-warning", "getdep-warning", "cancel-warning"}))
			Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
				ccv3.Query{Key: ccv3.NameFilter, Values: []string{"app-name"}},
				ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"space-guid"}},
			))
			Expect(fakeCloudControllerClient.GetDeploymentsArgsForCall(0)).To(ConsistOf(
				ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{"app-guid"}},
				ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
				ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.CreatedAtDescendingOrder}},
			))
			Expect(fakeCloudControllerClient.CancelDeploymentArgsForCall(0)).To(Equal("deployment-guid"))
		})

		Context("when no deployments are found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDeploymentsReturns([]resources.Deployment{}, nil, nil)
			})

			It("errors appropriately", func() {
				_, err := actor.CancelDeploymentByAppNameAndSpace("app-name", "space-guid")
				Expect(err).To(MatchError("failed to find a deployment for that app"))
			})
		})

		Context("when we fail while searching for app", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(nil, nil, errors.New("banana"))
			})

			It("errors appropriately", func() {
				_, err := actor.CancelDeploymentByAppNameAndSpace("app-name", "space-guid")
				Expect(err).To(MatchError("banana"))
			})
		})

		Context("when we fail while searching for the apps current deployment", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDeploymentsReturns(nil, nil, errors.New("vegetable"))
			})

			It("errors appropriately", func() {
				_, err := actor.CancelDeploymentByAppNameAndSpace("app-name", "space-guid")
				Expect(err).To(MatchError("vegetable"))
			})
		})
	})

	Describe("CreateApplicationDeployment", func() {

		Context("When there is no error", func() {

			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationDeploymentReturns("some-deployment-guid", ccv3.Warnings{"create-deployment-warning"}, nil)
			})

			It("Returns the deployment GUID when it is non empty", func() {
				deploymentGUID, warnings, err := actor.CreateDeployment("some-app-guid", "some-droplet-guid")
				Expect(deploymentGUID).To(Equal("some-deployment-guid"))
				Expect(warnings).To(ConsistOf("create-deployment-warning"))
				Expect(err).To(BeNil())
			})
		})

		Context("When an error occurs", func() {

			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationDeploymentReturns("", ccv3.Warnings{"create-deployment-warning"}, errors.New("failed create"))
			})

			It("Returns an error if an error occurred", func() {
				deploymentGUID, warnings, err := actor.CreateDeployment("some-app-guid", "some-droplet-guid")
				Expect(deploymentGUID).To(Equal(""))
				Expect(warnings).To(ConsistOf("create-deployment-warning"))
				Expect(err).To(MatchError(errors.New("failed create")))
			})
		})

	})

	Describe("GetDeploymentState", func() {

		Context("when there is no error", func() {

			BeforeEach(func() {
				resultDeployment := resources.Deployment{State: constant.DeploymentDeploying}
				fakeCloudControllerClient.GetDeploymentReturns(resultDeployment, ccv3.Warnings{"create-deployment-warning"}, nil)
			})

			It("returns a state of X", func() {
				deploymentState, warnings, err := actor.GetDeploymentState("some-deployment-guid")
				Expect(deploymentState).To(Equal(constant.DeploymentDeploying))
				Expect(warnings).To(ConsistOf("create-deployment-warning"))
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("PollDeployment", func() {
		var warningsChannel chan Warnings
		var allWarnings Warnings
		var funcDone chan interface{}

		BeforeEach(func() {
			fakeConfig.StartupTimeoutReturns(time.Second)
			fakeConfig.PollingIntervalReturns(0)
			warningsChannel = make(chan Warnings)
			allWarnings = Warnings{}
			funcDone = make(chan interface{})
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

		const myDeploymentGUID = "another-great-deployment-guid"

		Context("When the deployment eventually deploys", func() {
			BeforeEach(func() {
				getDeploymentCallCount := 0

				fakeCloudControllerClient.GetDeploymentStub = func(deploymentGuid string) (resources.Deployment, ccv3.Warnings, error) {
					defer func() { getDeploymentCallCount++ }()
					if getDeploymentCallCount == 0 {
						return resources.Deployment{State: constant.DeploymentDeploying},
							ccv3.Warnings{"get-process-warning-1", "get-process-warning-2"},
							nil
					} else {
						return resources.Deployment{State: constant.DeploymentDeployed},
							ccv3.Warnings{fmt.Sprintf("get-process-warning-%d", getDeploymentCallCount+2)},
							nil
					}
				}
			})

			It("returns a nil error", func() {
				err := actor.PollDeployment(myDeploymentGUID, warningsChannel)
				funcDone <- nil
				Expect(err).To(BeNil())
				Expect(allWarnings).To(ConsistOf("get-process-warning-1", "get-process-warning-2", "get-process-warning-3"))
			})
		})
		Context("When the deployment is cancelled", func() {
			BeforeEach(func() {
				getDeploymentCallCount := 0

				fakeCloudControllerClient.GetDeploymentStub = func(deploymentGuid string) (resources.Deployment, ccv3.Warnings, error) {
					defer func() { getDeploymentCallCount++ }()
					if getDeploymentCallCount == 0 {
						return resources.Deployment{State: constant.DeploymentDeploying},
							ccv3.Warnings{"get-process-warning-1", "get-process-warning-2"},
							nil
					} else {
						return resources.Deployment{State: constant.DeploymentCanceled},
							ccv3.Warnings{fmt.Sprintf("get-process-warning-%d", getDeploymentCallCount+2)},
							nil
					}
				}
			})
			It("throws a deployment canceled error", func() {
				err := actor.PollDeployment(myDeploymentGUID, warningsChannel)
				funcDone <- nil
				Expect(err).To(MatchError(errors.New("Deployment has been canceled")))
				Expect(allWarnings).To(ConsistOf("get-process-warning-1", "get-process-warning-2", "get-process-warning-3"))
			})

		})

		Context("When waiting for the deployment to finish times out", func() {
			BeforeEach(func() {
				fakeConfig.StartupTimeoutReturns(time.Millisecond)
				fakeConfig.PollingIntervalReturns(time.Millisecond * 2)
				fakeCloudControllerClient.GetDeploymentReturns(resources.Deployment{State: constant.DeploymentDeploying}, ccv3.Warnings{"some-deployment-warning"}, nil)
			})

			It("Throws a timeout error", func() {
				err := actor.PollDeployment(myDeploymentGUID, warningsChannel)
				funcDone <- nil
				Expect(err).To(MatchError(actionerror.StartupTimeoutError{}))
				Expect(allWarnings).To(ConsistOf("some-deployment-warning"))
			})
		})
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
			var processes []resources.Process

			BeforeEach(func() {
				fakeConfig.StartupTimeoutReturns(time.Second)
				fakeConfig.PollingIntervalReturns(0)
				processes = []resources.Process{
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
					err := actor.ZeroDowntimePollStart("some-guid", warningsChannel)
					funcDone <- nil

					Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
					Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
					Expect(err).To(MatchError(actionerror.StartupTimeoutError{}))
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
						processes = []resources.Process{
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
