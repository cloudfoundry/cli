package v3_test

import (
	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("allow-network-access Command", func() {
	var (
		cmd             AllowNetworkAccessCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeAllowNetworkAccessActor
		binaryName      string
		executeErr      error
		srcApp          string
		destApp         string
		protocol        string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeAllowNetworkAccessActor)

		srcApp = "some-app"
		destApp = "some-other-app"
		protocol = "tcp"

		cmd = AllowNetworkAccessCommand{
			UI:             testUI,
			Config:         fakeConfig,
			SharedActor:    fakeSharedActor,
			Actor:          fakeActor,
			RequiredArgs:   flag.AllowNetworkAccessArgs{SourceApp: srcApp},
			DestinationApp: destApp,
			Protocol:       flag.NetworkProtocol{Protocol: protocol},
			Port:           flag.NetworkPort{StartPort: 8080, EndPort: 8081},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(sharedaction.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			passedConfig, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(passedConfig).To(Equal(fakeConfig))
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		})

		It("outputs flavor text", func() {
			Expect(testUI.Out).To(Say("Allowing network traffic from app %s to %s in org some-org / space some-space as some-user...", srcApp, destApp))
		})

		Context("when the policy creation is successful", func() {
			BeforeEach(func() {
				fakeActor.AllowNetworkAccessReturns(cfnetworkingaction.Warnings{"some-warning-1", "some-warning-2"}, nil)
			})

			It("displays OK when no error occurs", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.AllowNetworkAccessCallCount()).To(Equal(1))
				passedSpaceGuid, passedSrcAppName, passedDestAppName, passedProtocol, passedStartPort, passedEndPort := fakeActor.AllowNetworkAccessArgsForCall(0)
				Expect(passedSpaceGuid).To(Equal("some-space-guid"))
				Expect(passedSrcAppName).To(Equal("some-app"))
				Expect(passedDestAppName).To(Equal("some-other-app"))
				Expect(passedProtocol).To(Equal("tcp"))
				Expect(passedStartPort).To(Equal(8080))
				Expect(passedEndPort).To(Equal(8081))

				Expect(testUI.Out).To(Say("Allowing network traffic from app %s to %s in org some-org / space some-space as some-user...", srcApp, destApp))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})
		Context("when the policy creation is successful", func() {
			BeforeEach(func() {
				fakeActor.AllowNetworkAccessReturns(cfnetworkingaction.Warnings{"some-warning-1", "some-warning-2"}, v3action.ApplicationNotFoundError{Name: srcApp})
			})

			It("displays OK when no error occurs", func() {
				Expect(executeErr).To(MatchError(translatableerror.ApplicationNotFoundError{Name: srcApp}))

				Expect(testUI.Out).To(Say("Allowing network traffic from app %s to %s in org some-org / space some-space as some-user...", srcApp, destApp))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
				Expect(testUI.Out).ToNot(Say("OK"))
			})
		})
	})
})
