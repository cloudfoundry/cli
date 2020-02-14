package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("space-ssh-allowed command", func() {
	var (
		cmd             SpaceSSHAllowedCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
		binaryName      string

		spaceName       string
		spaceSSHWarning v7action.Warnings
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		spaceName = RandomString("space")

		cmd = SpaceSSHAllowedCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.Space{Space: spaceName},
		}

		spaceSSHWarning = v7action.Warnings{"space-ssh-warning"}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: "some-org-guid"})

		fakeActor.GetSpaceFeatureReturns(true, spaceSSHWarning, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the environment is not set up correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("there is an error getting the SSH feature", func() {
		BeforeEach(func() {
			fakeActor.GetSpaceFeatureReturns(true, spaceSSHWarning, errors.New("some-feature-error"))
		})
		It("returns the error", func() {
			Expect(executeErr).To(HaveOccurred())
			Expect(executeErr).To(MatchError("some-feature-error"))
			Expect(testUI.Err).To(Say("space-ssh-warning"))
		})
	})

	It("prints the feature", func() {
		Expect(executeErr).To(Not(HaveOccurred()))
		Expect(fakeActor.GetSpaceFeatureCallCount()).To(Equal(1))
		inputSpaceName, inputOrgGUID, inputFeature := fakeActor.GetSpaceFeatureArgsForCall(0)
		Expect(inputSpaceName).To(Equal(spaceName))
		Expect(inputOrgGUID).To(Equal("some-org-guid"))
		Expect(inputFeature).To(Equal("ssh"))

		Expect(testUI.Err).To(Say("space-ssh-warning"))
		Expect(testUI.Out).To(Say("ssh support is enabled in space '%s'.", spaceName))
	})
})
