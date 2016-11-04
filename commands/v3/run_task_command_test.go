package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actors/v3actions"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/commands/commandsfakes"
	"code.cloudfoundry.org/cli/commands/v3"
	"code.cloudfoundry.org/cli/commands/v3/common"
	"code.cloudfoundry.org/cli/commands/v3/v3fakes"
	"code.cloudfoundry.org/cli/utils/configv3"
	"code.cloudfoundry.org/cli/utils/ui"

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

	Context("when the user is not logged in", func() {
		It("returns a NotLoggedInError", func() {
			Expect(executeErr).To(MatchError(common.NotLoggedInError{}))
		})
	})

	Context("when an organization is not targetted", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
		})

		It("returns a NoTargetedOrgError", func() {
			Expect(executeErr).To(MatchError(common.NoTargetedOrgError{}))
		})
	})

	Context("when a space is not targetted", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
		})

		It("returns a NoTargetedSpaceError", func() {
			Expect(executeErr).To(MatchError(common.NoTargetedSpaceError{}))
		})
	})

	Context("when the user is logged in, a space and an org is targetted", func() {
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
		})

		Context("when getting the logged in user results in an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("got bananapants??")
				fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
			})

			It("returns the same error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		Context("when provided a valid application name", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{
					Name: "some-user",
				}, nil)

				cmd.RequiredArgs.AppName = "some-app-name"
				cmd.RequiredArgs.Command = "fake command"

				fakeActor.GetApplicationByNameAndSpaceReturns(
					v3actions.Application{GUID: "some-app-guid"},
					v3actions.Warnings{
						"get-application-warning-1",
						"get-application-warning-2",
					}, nil)
				fakeActor.RunTaskReturns(v3actions.Task{SequenceID: 3},
					v3actions.Warnings{
						"get-application-warning-3",
					}, nil)
			})

			It("runs a new task and outputs all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
				appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal("some-app-name"))
				Expect(spaceGUID).To(Equal("some-space-guid"))

				Expect(fakeActor.RunTaskCallCount()).To(Equal(1))
				appGUID, command := fakeActor.RunTaskArgsForCall(0)
				Expect(appGUID).To(Equal("some-app-guid"))
				Expect(command).To(Equal("fake command"))

				Expect(fakeUI.Out).To(Say(`get-application-warning-1
get-application-warning-2
Creating task for app some-app-name in org some-org / space some-space as some-user...
get-application-warning-3
OK

Task 3 has been submitted successfully for execution.`,
				))
			})
		})

		Context("when there are errors", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{
					Name: "some-user",
				}, nil)
			})

			Context("when a translated error is returned", func() {
				Context("when GetApplicationByNameAndSpace returns a translatable error", func() {
					var (
						returnedError error
						expectedErr   error
					)

					BeforeEach(func() {
						expectedErr = errors.New("request-error")
						returnedError = cloudcontroller.RequestError{
							Err: expectedErr,
						}
						fakeActor.GetApplicationByNameAndSpaceReturns(
							v3actions.Application{GUID: "some-app-guid"},
							nil,
							returnedError)
					})

					It("returns a translatable error", func() {
						Expect(executeErr).To(MatchError(common.APIRequestError{Err: expectedErr}))
					})
				})

				Context("when RunTask returns a translatable error", func() {
					var returnedError error

					BeforeEach(func() {
						returnedError = cloudcontroller.UnverifiedServerError{URL: "some-url"}
						fakeActor.GetApplicationByNameAndSpaceReturns(
							v3actions.Application{GUID: "some-app-guid"},
							nil,
							nil)
						fakeActor.RunTaskReturns(
							v3actions.Task{},
							nil,
							returnedError)
					})

					It("returns a translatable error", func() {
						Expect(executeErr).To(MatchError(common.InvalidSSLCertError{API: "some-url"}))
					})
				})
			})

			Context("when an untranslatable error is returned", func() {
				Context("when GetApplicationByNameAndSpace returns an error", func() {
					var returnedError error

					BeforeEach(func() {
						returnedError = errors.New("got bananapants??")
						fakeActor.GetApplicationByNameAndSpaceReturns(v3actions.Application{GUID: "some-app-guid"},
							v3actions.Warnings{
								"get-application-warning-1",
								"get-application-warning-2",
							}, returnedError)
					})

					It("return the same error and outputs the warnings", func() {
						Expect(executeErr).To(MatchError(returnedError))

						Expect(fakeUI.Out).To(Say("get-application-warning-1"))
						Expect(fakeUI.Out).To(Say("get-application-warning-2"))
					})
				})

				Context("when RunTask returns an error", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("got bananapants??")
						fakeActor.GetApplicationByNameAndSpaceReturns(v3actions.Application{GUID: "some-app-guid"},
							v3actions.Warnings{
								"get-application-warning-1",
								"get-application-warning-2",
							}, nil)
						fakeActor.RunTaskReturns(v3actions.Task{},
							v3actions.Warnings{
								"run-task-warning-1",
								"run-task-warning-2",
							}, expectedErr)
					})

					It("returns the same error and outputs all warnings", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(fakeUI.Out).To(Say("get-application-warning-1"))
						Expect(fakeUI.Out).To(Say("get-application-warning-2"))
						Expect(fakeUI.Out).To(Say("run-task-warning-1"))
						Expect(fakeUI.Out).To(Say("run-task-warning-2"))
					})
				})
			})
		})
	})
})
