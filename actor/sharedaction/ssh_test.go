package sharedaction_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	"code.cloudfoundry.org/cli/util/clissh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSH Actions", func() {
	var (
		fakeConfig            *sharedactionfakes.FakeConfig
		actor                 *Actor
		fakeSecureShellClient *sharedactionfakes.FakeSecureShellClient
	)

	BeforeEach(func() {
		fakeSecureShellClient = new(sharedactionfakes.FakeSecureShellClient)
		fakeConfig = new(sharedactionfakes.FakeConfig)
		actor = NewActor(fakeConfig)
	})

	Describe("ExecuteSecureShell", func() {
		var (
			sshOptions SSHOptions
			executeErr error
		)

		BeforeEach(func() {
			sshOptions = SSHOptions{
				Username:           "some-user",
				Passcode:           "some-passcode",
				Endpoint:           "some-endpoint",
				HostKeyFingerprint: "some-fingerprint",
				SkipHostValidation: true,
			}
		})

		JustBeforeEach(func() {
			executeErr = actor.ExecuteSecureShell(fakeSecureShellClient, sshOptions)
		})

		It("calls connect with the provided authorization info", func() {
			Expect(fakeSecureShellClient.ConnectCallCount()).To(Equal(1))
			usernameArg, passcodeArg, endpointArg, fingerprintArg, skipHostValidationArg := fakeSecureShellClient.ConnectArgsForCall(0)
			Expect(usernameArg).To(Equal("some-user"))
			Expect(passcodeArg).To(Equal("some-passcode"))
			Expect(endpointArg).To(Equal("some-endpoint"))
			Expect(fingerprintArg).To(Equal("some-fingerprint"))
			Expect(skipHostValidationArg).To(BeTrue())
		})

		Context("when connecting fails", func() {
			BeforeEach(func() {
				fakeSecureShellClient.ConnectReturns(errors.New("some-connect-error"))
			})

			It("does not call Close", func() {
				Expect(fakeSecureShellClient.CloseCallCount()).To(Equal(0))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("some-connect-error"))
			})
		})

		Context("when connecting succeeds", func() {
			BeforeEach(func() {
				sshOptions.LocalPortForwardSpecs = []LocalPortForward{
					{LocalAddress: "local-address-1", RemoteAddress: "remote-address-1"},
					{LocalAddress: "local-address-2", RemoteAddress: "remote-address-2"},
				}
			})

			AfterEach(func() {
				Expect(fakeSecureShellClient.CloseCallCount()).To(Equal(1))
			})

			It("forwards the local ports", func() {
				Expect(fakeSecureShellClient.LocalPortForwardCallCount()).To(Equal(1))
				Expect(fakeSecureShellClient.LocalPortForwardArgsForCall(0)).To(Equal(
					[]clissh.LocalPortForward{
						{LocalAddress: "local-address-1", RemoteAddress: "remote-address-1"},
						{LocalAddress: "local-address-2", RemoteAddress: "remote-address-2"},
					},
				))
			})

			Context("when local port forwarding fails", func() {
				BeforeEach(func() {
					fakeSecureShellClient.LocalPortForwardReturns(errors.New("some-forwarding-error"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("some-forwarding-error"))
				})
			})

			Context("when local port forwarding succeeds", func() {
				Context("when skipping remote execution", func() {
					BeforeEach(func() {
						sshOptions.SkipRemoteExecution = true
					})

					It("waits and does not create an interactive session", func() {
						Expect(fakeSecureShellClient.WaitCallCount()).To(Equal(1))
						Expect(fakeSecureShellClient.InteractiveSessionCallCount()).To(Equal(0))
					})

					Context("when waiting errors", func() {
						// TODO: Handle different errors caused by interrupt signals
						BeforeEach(func() {
							fakeSecureShellClient.WaitReturns(errors.New("some-wait-error"))
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError(errors.New("some-wait-error")))
						})
					})

					Context("when waiting succeeds", func() {
						It("returns no error", func() {
							Expect(executeErr).ToNot(HaveOccurred())
						})
					})
				})

				Context("when creating an interactive session", func() {
					BeforeEach(func() {
						sshOptions.SkipRemoteExecution = false
						sshOptions.Commands = []string{"some-command-1", "some-command-2"}
						sshOptions.TTYOption = RequestTTYForce
					})

					It("creates an interactive session with the provided commands and tty type", func() {
						Expect(fakeSecureShellClient.InteractiveSessionCallCount()).To(Equal(1))
						commandsArg, ttyTypeArg := fakeSecureShellClient.InteractiveSessionArgsForCall(0)
						Expect(commandsArg).To(ConsistOf("some-command-1", "some-command-2"))
						Expect(ttyTypeArg).To(Equal(clissh.RequestTTYForce))
						Expect(fakeSecureShellClient.WaitCallCount()).To(Equal(0))
					})

					Context("when the interactive session errors", func() {
						// TODO: Handle different errors caused by interrupt signals
						BeforeEach(func() {
							fakeSecureShellClient.InteractiveSessionReturns(errors.New("some-interactive-session-error"))
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError("some-interactive-session-error"))
						})
					})

					Context("when the interactive session succeeds", func() {
						It("returns no error", func() {
							Expect(executeErr).ToNot(HaveOccurred())
						})
					})
				})
			})
		})
	})
})
