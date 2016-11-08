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

var _ = Describe("Tasks Command", func() {
	var (
		cmd        v3.TasksCommand
		fakeUI     *ui.UI
		fakeActor  *v3fakes.FakeTasksActor
		fakeConfig *commandsfakes.FakeConfig
		executeErr error
	)

	BeforeEach(func() {
		out := NewBuffer()
		fakeUI = ui.NewTestUI(nil, out, out)
		fakeActor = new(v3fakes.FakeTasksActor)
		fakeConfig = new(commandsfakes.FakeConfig)
		fakeConfig.ExperimentalReturns(true)

		cmd = v3.TasksCommand{
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

	Context("when the user is logged in, and a space and an org is targetted", func() {
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
				expectedErr = errors.New("get current user error")
				fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
			})

			It("returns the same error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		Context("when getting the logged in user does not result in an error", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{
					Name: "some-user",
				}, nil)
			})

			Context("when provided a valid application name", func() {
				BeforeEach(func() {
					cmd.RequiredArgs.AppName = "some-app-name"

					fakeActor.GetApplicationByNameAndSpaceReturns(
						v3actions.Application{GUID: "some-app-guid"},
						v3actions.Warnings{
							"get-application-warning-1",
							"get-application-warning-2",
						}, nil)
					fakeActor.GetApplicationTasksReturns(
						[]v3actions.Task{
							{
								GUID:       "task-3-guid",
								SequenceID: 3,
								Name:       "task-3",
								State:      "RUNNING",
								CreatedAt:  "some-time",
								Command:    "some-command",
							},
							{
								GUID:       "task-2-guid",
								SequenceID: 2,
								Name:       "task-2",
								State:      "FAILED",
								CreatedAt:  "some-time",
								Command:    "some-command",
							},
							{
								GUID:       "task-1-guid",
								SequenceID: 1,
								Name:       "task-1",
								State:      "SUCCEEDED",
								CreatedAt:  "some-time",
								Command:    "some-command",
							},
						},
						v3actions.Warnings{
							"get-tasks-warning-1",
						}, nil)
				})

				It("outputs all tasks associated with the application and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app-name"))
					Expect(spaceGUID).To(Equal("some-space-guid"))

					Expect(fakeActor.GetApplicationTasksCallCount()).To(Equal(1))
					guid, order := fakeActor.GetApplicationTasksArgsForCall(0)
					Expect(guid).To(Equal("some-app-guid"))
					Expect(order).To(Equal(v3actions.Descending))

					Expect(fakeUI.Out).To(Say(`get-application-warning-1
get-application-warning-2
Getting tasks for app some-app-name in org some-org / space some-space as some-user...
get-tasks-warning-1
OK

id   name     state       start time   command
3    task-3   RUNNING     some-time    some-command
2    task-2   FAILED      some-time    some-command
1    task-1   SUCCEEDED   some-time    some-command`,
					))
				})
			})

			Context("when there are errors", func() {
				Context("when a translated error is returned", func() {
					Context("when GetApplicationByNameAndSpace returns a translatable error", func() {
						var (
							returnedErr error
							expectedErr error
						)

						BeforeEach(func() {
							expectedErr = errors.New("request-error")
							returnedErr = cloudcontroller.RequestError{
								Err: expectedErr,
							}
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3actions.Application{GUID: "some-app-guid"},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(common.APIRequestError{Err: expectedErr}))
						})
					})

					Context("when GetApplicationTasks returns a translatable error", func() {
						var returnedErr error

						BeforeEach(func() {
							returnedErr = cloudcontroller.UnverifiedServerError{URL: "some-url"}
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3actions.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetApplicationTasksReturns(
								[]v3actions.Task{},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(common.InvalidSSLCertError{API: "some-url"}))
						})
					})
				})

				Context("when an untranslatable error is returned", func() {
					Context("when GetApplicationByNameAndSpace returns an error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("bananapants")
							fakeActor.GetApplicationByNameAndSpaceReturns(v3actions.Application{GUID: "some-app-guid"},
								v3actions.Warnings{
									"get-application-warning-1",
									"get-application-warning-2",
								}, expectedErr)
						})

						It("return the same error and outputs the warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(fakeUI.Out).To(Say("get-application-warning-1"))
							Expect(fakeUI.Out).To(Say("get-application-warning-2"))
						})
					})

					Context("when GetApplicationTasks returns an error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("bananapants??")
							fakeActor.GetApplicationByNameAndSpaceReturns(v3actions.Application{GUID: "some-app-guid"},
								v3actions.Warnings{
									"get-application-warning-1",
									"get-application-warning-2",
								}, nil)
							fakeActor.GetApplicationTasksReturns(nil,
								v3actions.Warnings{
									"get-tasks-warning-1",
									"get-tasks-warning-2",
								}, expectedErr)
						})

						It("returns the same error and outputs all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(fakeUI.Out).To(Say("get-application-warning-1"))
							Expect(fakeUI.Out).To(Say("get-application-warning-2"))
							Expect(fakeUI.Out).To(Say("get-tasks-warning-1"))
							Expect(fakeUI.Out).To(Say("get-tasks-warning-2"))
						})
					})
				})
			})
		})
	})
})
