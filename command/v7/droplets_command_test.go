package v7_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("droplets Command", func() {
	var (
		cmd             DropletsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = DropletsCommand{
			RequiredArgs: flag.AppName{AppName: "some-app"},
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				Actor:       fakeActor,
				SharedActor: fakeSharedActor,
			},
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})

		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("getting the application droplets returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = ccerror.RequestError{}
			fakeActor.GetApplicationDropletsReturns([]resources.Droplet{}, v7action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(ccerror.RequestError{}))

			Expect(testUI.Out).To(Say(`Getting droplets of app some-app in org some-org / space some-space as steve\.\.\.`))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("getting the application droplets returns some droplets", func() {
		var createdAtOne, createdAtTwo string
		BeforeEach(func() {
			createdAtOne = "2017-08-14T21:16:42Z"
			createdAtTwo = "2017-08-16T00:18:24Z"
			droplets := []resources.Droplet{
				{
					GUID:      "some-droplet-guid-1",
					State:     constant.DropletStaged,
					CreatedAt: createdAtOne,
					IsCurrent: true,
				},
				{
					GUID:      "some-droplet-guid-2",
					State:     constant.DropletFailed,
					CreatedAt: createdAtTwo,
					IsCurrent: false,
				},
			}
			fakeActor.GetApplicationDropletsReturns(droplets, v7action.Warnings{"warning-1", "warning-2"}, nil)
		})

		It("prints the application droplets and outputs warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Getting droplets of app some-app in org some-org / space some-space as steve\.\.\.\n`))
			Expect(testUI.Out).To(Say("\n"))

			createdAtOneParsed, err := time.Parse(time.RFC3339, createdAtOne)
			Expect(err).ToNot(HaveOccurred())
			createdAtTwoParsed, err := time.Parse(time.RFC3339, createdAtTwo)
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`guid\s+state\s+created\n`))
			Expect(testUI.Out).To(Say(`some-droplet-guid-1 \(current\)\s+staged\s+%s\n`, testUI.UserFriendlyDate(createdAtOneParsed)))
			Expect(testUI.Out).To(Say(`some-droplet-guid-2\s+failed\s+%s\n`, testUI.UserFriendlyDate(createdAtTwoParsed)))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))

			Expect(fakeActor.GetApplicationDropletsCallCount()).To(Equal(1))
			appName, spaceGUID := fakeActor.GetApplicationDropletsArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
		})
	})

	When("getting the application droplets returns no droplets", func() {
		BeforeEach(func() {
			fakeActor.GetApplicationDropletsReturns([]resources.Droplet{}, v7action.Warnings{"warning-1", "warning-2"}, nil)
		})

		It("displays there are no droplets", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Getting droplets of app some-app in org some-org / space some-space as steve\.\.\.`))
			Expect(testUI.Out).To(Say("No droplets found"))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})
})
