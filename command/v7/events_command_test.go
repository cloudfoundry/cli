package v7_test

import (
	"errors"
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/flag"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("events Command", func() {
	var (
		cmd             EventsCommand
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

		cmd = EventsCommand{
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

		fakeActor.GetCurrentUserReturns(configv3.User{Name: "steve"}, nil)
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
			fakeActor.GetCurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("getting the events returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = ccerror.RequestError{}
			fakeActor.GetRecentEventsByApplicationNameAndSpaceReturns(nil, v7action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(ccerror.RequestError{}))

			Expect(testUI.Out).To(Say(`Getting events for app some-app in org some-org / space some-space as steve\.\.\.`))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("getting the events returns some events", func() {
		var (
			createdAtOne, createdAtTwo time.Time
			errOne, errTwo             error
		)

		BeforeEach(func() {
			createdAtOne, errOne = time.Parse(time.RFC3339, "2017-08-14T21:16:42Z")
			createdAtTwo, errTwo = time.Parse(time.RFC3339, "2017-08-16T00:18:24Z")
			events := []v7action.Event{
				{
					GUID:      "some-event-guid-1",
					Type:      "audit.app.wow",
					ActorName: "user1",
					Time:      createdAtOne,
				},
				{
					GUID:        "some-event-guid-2",
					Type:        "audit.app.cool",
					ActorName:   "user2",
					Time:        createdAtTwo,
					Description: `"hello": "world"`,
				},
			}

			fakeActor.GetRecentEventsByApplicationNameAndSpaceReturns(events, v7action.Warnings{"warning-1", "warning-2"}, nil)
		})

		It("prints the events and outputs warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(errOne).ToNot(HaveOccurred())
			Expect(errTwo).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Getting events for app some-app in org some-org / space some-space as steve\.\.\.`))
			Expect(testUI.Out).To(Say("\n"))

			Expect(testUI.Out).To(Say(`time\s+event\s+actor\s+description`))
			Expect(testUI.Out).To(Say(`%s\s+audit.app.wow\s+user1\s+`, regexp.QuoteMeta(createdAtOne.Local().Format("2006-01-02T15:04:05.00-0700"))))
			Expect(testUI.Out).To(Say(`%s\s+audit.app.cool\s+user2\s+`, regexp.QuoteMeta(createdAtTwo.Local().Format("2006-01-02T15:04:05.00-0700"))))
			Expect(testUI.Out).To(Say(`"hello": "world"`))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))

			Expect(fakeActor.GetRecentEventsByApplicationNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID := fakeActor.GetRecentEventsByApplicationNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
		})
	})

	When("getting the application events returns no events", func() {
		BeforeEach(func() {
			fakeActor.GetRecentEventsByApplicationNameAndSpaceReturns([]v7action.Event{}, v7action.Warnings{"warning-1", "warning-2"}, nil)
		})

		It("displays there are no events", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Getting events for app some-app in org some-org / space some-space as steve\.\.\.`))
			Expect(testUI.Out).To(Say("No events found"))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})
})
