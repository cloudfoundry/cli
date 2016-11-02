package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actors/v3actions"
	"code.cloudfoundry.org/cli/commands/commandsfakes"
	"code.cloudfoundry.org/cli/commands/v2/common"
	"code.cloudfoundry.org/cli/commands/v3"
	"code.cloudfoundry.org/cli/utils/configv3"
	"code.cloudfoundry.org/cli/utils/ui"
	// 	"code.cloudfoundry.org/cli/actors/v2actions"
	// 	"code.cloudfoundry.org/cli/commands/commandsfakes"

	// 	"code.cloudfoundry.org/cli/commands/v2/common"
	"code.cloudfoundry.org/cli/commands/v3/v3fakes"
	// 	"code.cloudfoundry.org/cli/utils/configv3"
	// 	"code.cloudfoundry.org/cli/utils/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("RunTask Command", func() {
	var (
		cmd        v3.RunTaskCommand
		fakeUI     *ui.UI
		fakeActor  *v3fakes.FakeRunTaskActor
		fakeConfig *commandsfakes.FakeConfig
		executeErr error
	)

	BeforeEach(func() {
		out := NewBuffer()
		fakeUI = ui.NewTestUI(nil, out, out)
		fakeActor = new(v3fakes.FakeRunTaskActor)
		fakeConfig = new(commandsfakes.FakeConfig)
		fakeConfig.ExperimentalReturns(true)

		cmd = v3.RunTaskCommand{
			UI:     fakeUI,
			Actor:  fakeActor,
			Config: fakeConfig,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the target is not set", func() {
		It("returns an error", func() {
			Expect(executeErr).To(MatchError(common.NotLoggedInError{}))
		})
	})

	Context("when getting the logged in user results in an error", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
			expectedErr = errors.New("got bananapants??")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("returns the same error", func() {
			Expect(executeErr).To(MatchError(expectedErr))
		})
	})

	Context("when given a staged app", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{
				Name: "some-user",
			}, nil)

			cmd.RequiredArgs.AppName = "some-app-name"
			cmd.RequiredArgs.Command = "fake command"
			fakeActor.GetApplicationByNameAndSpaceReturns(v3actions.Application{GUID: "some-app-guid"}, nil, nil)
			fakeActor.RunTaskReturns(v3actions.Task{SequenceID: 3}, nil, nil)
		})

		It("runs a new task", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("some-app-name"))
			Expect(spaceGUID).To(Equal("some-space-guid"))

			Expect(fakeActor.RunTaskCallCount()).To(Equal(1))
			appGUID, command := fakeActor.RunTaskArgsForCall(0)
			Expect(appGUID).To(Equal("some-app-guid"))
			Expect(command).To(Equal("fake command"))

			Expect(fakeUI.Out).To(Say("Creating task for app some-app-name in org some-org / space some-space as some-user..."))
			Expect(fakeUI.Out).To(Say("OK"))
			Expect(fakeUI.Out).To(Say("Task 3 has been submitted successfully for execution."))
		})
	})

	Context("when there are warnings", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{
				Name: "some-user",
			}, nil)

		})

		Context("when GetApplicationByNameAndSpace returns warnings and an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("got bananapants??")
				fakeActor.GetApplicationByNameAndSpaceReturns(v3actions.Application{GUID: "some-app-guid"},
					v3actions.Warnings{
						"get-application-warning-1",
						"get-application-warning-2",
					}, expectedErr)
			})

			It("outputs the warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))

				Expect(fakeUI.Out).To(Say("get-application-warning-1"))
				Expect(fakeUI.Out).To(Say("get-application-warning-2"))
			})
		})

		Context("when RunTask returns warnings and an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("got bananapants??")
				fakeActor.RunTaskReturns(v3actions.Task{SequenceID: 3},
					v3actions.Warnings{
						"run-task-warning-1",
						"run-task-warning-2",
					}, expectedErr)
			})

			It("outputs the warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))

				Expect(fakeUI.Out).To(Say("run-task-warning-1"))
				Expect(fakeUI.Out).To(Say("run-task-warning-2"))
			})
		})

		Context("when multiple actor methods have warnings", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(v3actions.Application{GUID: "some-app-guid"},
					v3actions.Warnings{
						"get-application-warning",
					}, nil)
				fakeActor.RunTaskReturns(v3actions.Task{SequenceID: 3},
					v3actions.Warnings{
						"run-task-warning",
					}, nil)
			})
			It("outputs all the warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeUI.Out).To(Say("get-application-warning"))
				Expect(fakeUI.Out).To(Say("run-task-warning"))
			})
		})
	})

	Context("when given an unstaged app", func() {
		It("fails", func() {
			//fake out actor error
			//returns error code 1
		})
	})

	Context("when given a non-existent app", func() {
		It("fails and returns an 'app not found' error", func() {
		})
	})
})
