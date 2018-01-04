package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSH Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
		fakeConfig                *v3actionfakes.FakeConfig
		fakeSharedActor           *v3actionfakes.FakeSharedActor
		fakeUAAClient             *v3actionfakes.FakeUAAClient
		executeErr                error
		warnings                  Warnings
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v3actionfakes.FakeConfig)
		fakeSharedActor = new(v3actionfakes.FakeSharedActor)
		fakeUAAClient = new(v3actionfakes.FakeUAAClient)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, fakeUAAClient)
	})

	Describe("GetSecureShellConfigurationByApplicationNameSpaceProcessTypeAndIndex", func() {
		var sshAuth SSHAuthentication

		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.SSHOAuthClientReturns("some-access-oauth-client")
		})

		JustBeforeEach(func() {
			sshAuth, warnings, executeErr = actor.GetSecureShellConfigurationByApplicationNameSpaceProcessTypeAndIndex("some-app", "some-space-guid", "some-process-type", 0)
		})

		Context("when the app ssh endpoint is empty", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.AppSSHEndpointReturns("")
			})
			It("creates an ssh-endpoint-not-set error", func() {
				Expect(executeErr).To(MatchError("SSH endpoint not set"))
			})
		})

		Context("when the app ssh hostkey fingerprint is empty", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.AppSSHEndpointReturns("some-app-ssh-endpoint")
				fakeCloudControllerClient.AppSSHHostKeyFingerprintReturns("")
			})
			It("creates an ssh-hostkey-fingerprint-not-set error", func() {
				Expect(executeErr).To(MatchError("SSH hostkey fingerprint not set"))
			})
		})

		Context("when ssh endpoint and fingerprint are set", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.AppSSHEndpointReturns("some-app-ssh-endpoint")
				fakeCloudControllerClient.AppSSHHostKeyFingerprintReturns("some-app-ssh-fingerprint")
			})

			It("looks up the passcode with the config credentials", func() {
				Expect(fakeUAAClient.GetSSHPasscodeCallCount()).To(Equal(1))
				accessTokenArg, oathClientArg := fakeUAAClient.GetSSHPasscodeArgsForCall(0)
				Expect(accessTokenArg).To(Equal("some-access-token"))
				Expect(oathClientArg).To(Equal("some-access-oauth-client"))
			})

			Context("when getting the ssh passcode errors", func() {
				BeforeEach(func() {
					fakeUAAClient.GetSSHPasscodeReturns("", errors.New("some-ssh-passcode-error"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("some-ssh-passcode-error"))
				})
			})

			Context("when getting the ssh passcode succeeds", func() {
				BeforeEach(func() {
					fakeUAAClient.GetSSHPasscodeReturns("some-ssh-passcode", nil)
				})

				Context("when getting the application summary errors", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationsReturns(nil, ccv3.Warnings{"some-app-warnings"}, errors.New("some-application-summary-error"))
					})

					It("returns all warnings and the error", func() {
						Expect(executeErr).To(MatchError("some-application-summary-error"))
						Expect(warnings).To(ConsistOf("some-app-warnings"))
					})
				})

				Context("when getting the application summary succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{Name: "some-app"}}, ccv3.Warnings{"some-app-warnings"}, nil)
					})

					Context("when the process does not exist", func() {
						It("returns all warnings and the error", func() {
							Expect(executeErr).To(MatchError(actionerror.ProcessNotFoundError{ProcessType: "some-process-type"}))
							Expect(warnings).To(ConsistOf("some-app-warnings"))
						})
					})

					Context("when the application is not in the STARTED state", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationProcessesReturns([]ccv3.Process{{Type: "some-process-type", GUID: "some-process-guid"}}, ccv3.Warnings{"some-process-warnings"}, nil)
						})

						It("returns a ApplicationNotStartedError", func() {
							Expect(executeErr).To(MatchError(actionerror.ApplicationNotStartedError{Name: "some-app"}))
							Expect(warnings).To(ConsistOf("some-app-warnings", "some-process-warnings"))
						})
					})

					Context("when the process doesn't have the specified instance index", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{Name: "some-app", State: constant.ApplicationStarted}}, ccv3.Warnings{"some-app-warnings"}, nil)
							fakeCloudControllerClient.GetApplicationProcessesReturns([]ccv3.Process{{Type: "some-process-type", GUID: "some-process-guid"}}, ccv3.Warnings{"some-process-warnings"}, nil)
						})

						It("returns a ProcessIndexNotFoundError", func() {
							Expect(executeErr).To(MatchError(actionerror.ProcessInstanceNotFoundError{ProcessType: "some-process-type", InstanceIndex: 0}))
						})
					})

					Context("when the process instance is not RUNNING", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{Name: "some-app", State: constant.ApplicationStarted}}, ccv3.Warnings{"some-app-warnings"}, nil)
							fakeCloudControllerClient.GetApplicationProcessesReturns([]ccv3.Process{{Type: "some-process-type", GUID: "some-process-guid"}}, ccv3.Warnings{"some-process-warnings"}, nil)
							fakeCloudControllerClient.GetProcessInstancesReturns([]ccv3.ProcessInstance{{State: constant.ProcessInstanceDown, Index: 0}}, ccv3.Warnings{"some-instance-warnings"}, nil)
						})
						It("returns a ProcessInstanceNotRunningError", func() {
							Expect(executeErr).To(MatchError(actionerror.ProcessInstanceNotRunningError{ProcessType: "some-process-type", InstanceIndex: 0}))
						})
					})

					Context("when the specified process and index exist, app is STARTED and the instance is RUNNING", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{Name: "some-app", State: constant.ApplicationStarted}}, ccv3.Warnings{"some-app-warnings"}, nil)
							fakeCloudControllerClient.GetApplicationProcessesReturns([]ccv3.Process{{Type: "some-process-type", GUID: "some-process-guid"}}, ccv3.Warnings{"some-process-warnings"}, nil)
							fakeCloudControllerClient.GetProcessInstancesReturns([]ccv3.ProcessInstance{{State: constant.ProcessInstanceRunning, Index: 0}}, ccv3.Warnings{"some-instance-warnings"}, nil)
						})

						Context("when starting the secure session succeeds", func() {
							It("returns all warnings", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(warnings).To(ConsistOf("some-app-warnings", "some-process-warnings", "some-instance-warnings"))

								Expect(sshAuth).To(Equal(SSHAuthentication{
									Endpoint:           "some-app-ssh-endpoint",
									HostKeyFingerprint: "some-app-ssh-fingerprint",
									Passcode:           "some-ssh-passcode",
									Username:           "cf:some-process-guid/0",
								}))

								Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
								Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
									ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app"}},
									ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
								))
							})
						})
					})
				})
			})
		})
	})
})
