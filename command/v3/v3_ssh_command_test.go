package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-ssh Command", func() {
	var (
		cmd             v3.V3SSHCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3SSHActor
		executeErr      error
		appName         string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3SSHActor)

		appName = "some-app"
		cmd = v3.V3SSHCommand{
			RequiredArgs: flag.AppName{AppName: appName},

			ProcessType:         "some-process-type",
			ProcessIndex:        1,
			Commands:            []string{"some", "commands"},
			SkipHostValidation:  true,
			SkipRemoteExecution: true,

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionV3)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the API version is below the minimum", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns("0.0.0")
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumAPIVersionNotMetError{
				CurrentVersion: "0.0.0",
				MinimumVersion: ccversion.MinVersionV3,
			}))
		})

		It("displays the experimental warning", func() {
			Expect(testUI.Out).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		})
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: "steve"})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.NotLoggedInError{BinaryName: "steve"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is targeted to an organization and space", func() {
		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: "some-space-guid"})
		})

		Context("when executing the secure shell fails", func() {
			BeforeEach(func() {
				fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexReturns(v3action.Warnings{"some-warnings"}, errors.New("some-error"))
			})

			It("returns the error and displays all warnings", func() {
				Expect(executeErr).To(MatchError(errors.New("some-error")))
				Expect(testUI.Err).To(Say("some-warnings"))
			})
		})

		Context("when executing the secure shell succeeds", func() {
			BeforeEach(func() {
				fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexReturns(v3action.Warnings{"some-warnings"}, nil)
				cmd.DisablePseudoTTY = true
			})

			It("returns nil and displays all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warnings"))

				Expect(fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexCallCount()).To(Equal(1))
				appNameArg, spaceGUIDArg, processTypeArg, processIndexArg, sshOptionsArg := fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexArgsForCall(0)
				Expect(appNameArg).To(Equal(appName))
				Expect(spaceGUIDArg).To(Equal("some-space-guid"))
				Expect(processTypeArg).To(Equal("some-process-type"))
				Expect(processIndexArg).To(Equal(uint(1)))
				Expect(sshOptionsArg).To(Equal(v3action.SSHOptions{
					Commands:            []string{"some", "commands"},
					TTYOption:           sharedaction.RequestTTYNo,
					SkipHostValidation:  true,
					SkipRemoteExecution: true,
				}))
			})

			Context("when the user doesn't provide a process-type and index", func() {
				BeforeEach(func() {
					cmd.ProcessType = ""
					cmd.ProcessIndex = 0
				})

				It("defaults to 'web' and index 0", func() {
					Expect(fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexCallCount()).To(Equal(1))
					_, _, processTypeArg, processIndexArg, _ := fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexArgsForCall(0)
					Expect(processTypeArg).To(Equal("web"))
					Expect(processIndexArg).To(Equal(uint(0)))
				})
			})
		})

		Context("when working with local port forwarding", func() {
			BeforeEach(func() {
				cmd.LocalPortForwardSpecs = []flag.SSHPortForwarding{
					{LocalAddress: "localhost:8888", RemoteAddress: "remote:4444"},
					{LocalAddress: "localhost:7777", RemoteAddress: "remote:3333"},
				}

				fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexReturns(v3action.Warnings{"some-warnings"}, nil)
				cmd.DisablePseudoTTY = true
			})

			It("passes along port forwarding information", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warnings"))

				Expect(fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexCallCount()).To(Equal(1))
				appNameArg, spaceGUIDArg, processTypeArg, processIndexArg, sshOptionsArg := fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexArgsForCall(0)
				Expect(appNameArg).To(Equal(appName))
				Expect(spaceGUIDArg).To(Equal("some-space-guid"))
				Expect(processTypeArg).To(Equal("some-process-type"))
				Expect(processIndexArg).To(Equal(uint(1)))
				Expect(sshOptionsArg).To(Equal(v3action.SSHOptions{
					Commands:            []string{"some", "commands"},
					TTYOption:           sharedaction.RequestTTYNo,
					SkipHostValidation:  true,
					SkipRemoteExecution: true,
					LocalPortForwardSpecs: []sharedaction.LocalPortForward{
						{LocalAddress: "localhost:8888", RemoteAddress: "remote:4444"},
						{LocalAddress: "localhost:7777", RemoteAddress: "remote:3333"},
					},
				}))
			})
		})

		Context("when a tty flag is provided", func() {
			DescribeTable("tty combinations",
				func(disablePseudoTTY bool, forcePseudoTTY bool, requestPseudoTTY bool, expectedErr error, ttyOption sharedaction.TTYOption) {
					cmd.DisablePseudoTTY = disablePseudoTTY
					cmd.ForcePseudoTTY = forcePseudoTTY
					cmd.RequestPseudoTTY = requestPseudoTTY
					executeErr = cmd.Execute(nil)

					if expectedErr == nil {
						Expect(executeErr).To(BeNil())
						Expect(fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexCallCount()).To(Equal(2))
						_, _, _, _, sshOptionsArg := fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexArgsForCall(1)
						Expect(sshOptionsArg.TTYOption).To(Equal(ttyOption))
					} else {
						Expect(executeErr).To(MatchError(expectedErr))
					}
				},
				Entry("default", false, false, false, nil, sharedaction.RequestTTYAuto),
				Entry("disable tty", true, false, false, nil, sharedaction.RequestTTYNo),
				Entry("force tty", false, true, false, nil, sharedaction.RequestTTYForce),
				Entry("force tty", false, false, true, nil, sharedaction.RequestTTYYes),
				Entry("disable and force tty", true, true, false,
					translatableerror.ArgumentCombinationError{Args: []string{
						"--disable-pseudo-tty", "-T", "--force-pseudo-tty", "--request-pseudo-tty", "-t",
					}},
					sharedaction.TTYOption(0),
				),
				Entry("disable and requst tty", true, false, true,
					translatableerror.ArgumentCombinationError{Args: []string{
						"--disable-pseudo-tty", "-T", "--force-pseudo-tty", "--request-pseudo-tty", "-t",
					}},
					sharedaction.TTYOption(0),
				),
				Entry("force and request tty", false, true, true,
					translatableerror.ArgumentCombinationError{Args: []string{
						"--disable-pseudo-tty", "-T", "--force-pseudo-tty", "--request-pseudo-tty", "-t",
					}},
					sharedaction.TTYOption(0),
				),
				Entry("disable, force, and request tty", true, true, true,
					translatableerror.ArgumentCombinationError{Args: []string{
						"--disable-pseudo-tty", "-T", "--force-pseudo-tty", "--request-pseudo-tty", "-t",
					}},
					sharedaction.TTYOption(0),
				),
			)
		})
	})
})
