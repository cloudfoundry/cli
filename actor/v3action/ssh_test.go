package v3action_test

import (
	"errors"
	"net/url"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
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

		forwardSpecs []sharedaction.LocalPortForward
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v3actionfakes.FakeConfig)
		fakeSharedActor = new(v3actionfakes.FakeSharedActor)
		fakeUAAClient = new(v3actionfakes.FakeUAAClient)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, fakeUAAClient)

		forwardSpecs = []sharedaction.LocalPortForward{
			{LocalAddress: "localhost:8888", RemoteAddress: "remote:4444"},
			{LocalAddress: "localhost:7777", RemoteAddress: "remote:3333"},
		}
	})

	Describe("ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndex", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.SSHOAuthClientReturns("some-access-oauth-client")
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndex("some-app", "some-space-guid", "some-process-type", 0, SSHOptions{
				Commands:              []string{"some-command"},
				LocalPortForwardSpecs: forwardSpecs,
				TTYOption:             sharedaction.RequestTTYForce,
				SkipHostValidation:    true,
				SkipRemoteExecution:   true,
			})
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
						fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{Name: "some-app", State: "STARTED"}}, ccv3.Warnings{"some-app-warnings"}, nil)
						fakeCloudControllerClient.GetApplicationProcessesReturns([]ccv3.Process{{Type: "some-process-type", GUID: "some-process-guid", Instances: types.NullInt{IsSet: true, Value: 1}}}, ccv3.Warnings{"some-process-warnings"}, nil)
					})

					Context("when the process does not exist", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationProcessesReturns([]ccv3.Process{{Type: "some-other-type", GUID: "some-process-guid"}}, ccv3.Warnings{"some-process-warnings"}, nil)
						})

						It("returns all warnings and the error", func() {
							Expect(executeErr).To(MatchError(actionerror.ProcessNotFoundError{ProcessType: "some-process-type"}))
							Expect(warnings).To(ConsistOf("some-app-warnings", "some-process-warnings"))
						})
					})

					Context("when the application is not in the STARTED state", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{Name: "some-app", State: "STOPPED"}}, ccv3.Warnings{"some-app-warnings"}, nil)
						})

						It("returns a ApplicationNotStartedError", func() {
							Expect(executeErr).To(MatchError(actionerror.ApplicationNotStartedError{Name: "some-app"}))
						})
					})

					Context("when the process doesn't have the specified instance index", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationProcessesReturns([]ccv3.Process{{Type: "some-process-type", GUID: "some-process-guid"}}, ccv3.Warnings{"some-process-warnings"}, nil)
						})

						It("returns a ProcessIndexNotFoundError", func() {
							Expect(executeErr).To(MatchError(actionerror.ProcessInstanceNotFoundError{ProcessType: "some-process-type", InstanceIndex: 0}))
						})
					})

					Context("when the specified process and index exist and the applicaiton is STARTED", func() {
						Context("when starting the secure session fails", func() {
							BeforeEach(func() {
								fakeSharedActor.ExecuteSecureShellReturns(errors.New("some-ssh-connection-error"))
							})

							It("returns all warnings and the error", func() {
								Expect(executeErr).To(MatchError("some-ssh-connection-error"))
								Expect(warnings).To(ConsistOf("some-app-warnings", "some-process-warnings"))
							})
						})

						Context("when starting the secure session succeeds", func() {
							It("returns all warnings", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(warnings).To(ConsistOf("some-app-warnings", "some-process-warnings"))

								Expect(fakeSharedActor.ExecuteSecureShellCallCount()).To(Equal(1))
								Expect(fakeSharedActor.ExecuteSecureShellArgsForCall(0)).To(Equal(sharedaction.SSHOptions{
									Commands:              []string{"some-command"},
									Endpoint:              "some-app-ssh-endpoint",
									HostKeyFingerprint:    "some-app-ssh-fingerprint",
									LocalPortForwardSpecs: forwardSpecs,
									Passcode:              "some-ssh-passcode",
									SkipHostValidation:    true,
									SkipRemoteExecution:   true,
									TTYOption:             sharedaction.RequestTTYForce,
									Username:              "cf:some-process-guid/0",
								}))

								Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
								Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(Equal(url.Values{
									"space_guids": []string{"some-space-guid"},
									"names":       []string{"some-app"},
								}))
							})
						})
					})
				})
			})
		})
	})
})
