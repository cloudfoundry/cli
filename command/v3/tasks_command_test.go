package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Tasks Command", func() {
	var (
		cmd        v3.TasksCommand
		testUI     *ui.UI
		fakeActor  *v3fakes.FakeTasksActor
		fakeConfig *commandfakes.FakeConfig
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeActor = new(v3fakes.FakeTasksActor)
		fakeActor.CloudControllerAPIVersionReturns("3.0.0")
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.ExperimentalReturns(true)

		cmd = v3.TasksCommand{
			UI:     testUI,
			Actor:  fakeActor,
			Config: fakeConfig,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the API version is too small", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns("0.0.0")
		})

		It("returns a MinimumAPIVersionError", func() {
			Expect(executeErr).To(MatchError(command.MinimumAPIVersionError{
				CurrentVersion: "0.0.0",
				MinimumVersion: "3.0.0",
			}))
		})
	})

	Context("when the user is not logged in", func() {
		It("returns a NotLoggedInError", func() {
			Expect(executeErr).To(MatchError(command.NotLoggedInError{}))
		})
	})

	Context("when an organization is not targetted", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
		})

		It("returns a NoTargetedOrgError", func() {
			Expect(executeErr).To(MatchError(command.NoTargetedOrgError{}))
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
			Expect(executeErr).To(MatchError(command.NoTargetedSpaceError{}))
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
						v3action.Application{GUID: "some-app-guid"},
						v3action.Warnings{
							"get-application-warning-1",
							"get-application-warning-2",
						},
						nil,
					)
					fakeActor.GetApplicationTasksReturns(
						[]v3action.Task{
							{
								GUID:       "task-3-guid",
								SequenceID: 3,
								Name:       "task-3",
								State:      "RUNNING",
								CreatedAt:  "2016-11-08T22:26:02Z",
								Command:    "some-command",
							},
							{
								GUID:       "task-2-guid",
								SequenceID: 2,
								Name:       "task-2",
								State:      "FAILED",
								CreatedAt:  "2016-11-08T22:26:02Z",
								Command:    "some-command",
							},
							{
								GUID:       "task-1-guid",
								SequenceID: 1,
								Name:       "task-1",
								State:      "SUCCEEDED",
								CreatedAt:  "2016-11-08T22:26:02Z",
								Command:    "some-command",
							},
						},
						v3action.Warnings{
							"get-tasks-warning-1",
						},
						nil,
					)
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
					Expect(order).To(Equal(v3action.Descending))

					Expect(testUI.Out).To(Say(`Getting tasks for app some-app-name in org some-org / space some-space as some-user...
OK

id   name     state       start time                      command
3    task-3   RUNNING     Tue, 08 Nov 2016 22:26:02 UTC   some-command
2    task-2   FAILED      Tue, 08 Nov 2016 22:26:02 UTC   some-command
1    task-1   SUCCEEDED   Tue, 08 Nov 2016 22:26:02 UTC   some-command`,
					))
					Expect(testUI.Err).To(Say(`get-application-warning-1
get-application-warning-2
get-tasks-warning-1`))
				})

				Context("when the tasks' command fields are returned as empty strings", func() {
					BeforeEach(func() {
						fakeActor.GetApplicationTasksReturns(
							[]v3action.Task{
								{
									GUID:       "task-2-guid",
									SequenceID: 2,
									Name:       "task-2",
									State:      "FAILED",
									CreatedAt:  "2016-11-08T22:26:02Z",
									Command:    "",
								},
								{
									GUID:       "task-1-guid",
									SequenceID: 1,
									Name:       "task-1",
									State:      "SUCCEEDED",
									CreatedAt:  "2016-11-08T22:26:02Z",
									Command:    "",
								},
							},
							v3action.Warnings{
								"get-tasks-warning-1",
							},
							nil,
						)
					})

					It("outputs [hidden] for the tasks' commands", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`
2    task-2   FAILED      Tue, 08 Nov 2016 22:26:02 UTC   \[hidden\]
1    task-1   SUCCEEDED   Tue, 08 Nov 2016 22:26:02 UTC   \[hidden\]`,
						))
					})
				})

				Context("when there are no tasks associated with the application", func() {
					BeforeEach(func() {
						fakeActor.GetApplicationTasksReturns(
							[]v3action.Task{},
							nil,
							nil,
						)
					})

					It("outputs an empty table", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`
id   name   state   start time   command
`,
						))
						Expect(testUI.Out).NotTo(Say("1"))
					})
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
								v3action.Application{GUID: "some-app-guid"},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(command.APIRequestError{Err: expectedErr}))
						})
					})

					Context("when GetApplicationTasks returns a translatable error", func() {
						var returnedErr error

						BeforeEach(func() {
							returnedErr = cloudcontroller.UnverifiedServerError{URL: "some-url"}
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetApplicationTasksReturns(
								[]v3action.Task{},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(command.InvalidSSLCertError{API: "some-url"}))
						})
					})
				})

				Context("when an untranslatable error is returned", func() {
					Context("when GetApplicationByNameAndSpace returns an error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("bananapants")
							fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "some-app-guid"},
								v3action.Warnings{
									"get-application-warning-1",
									"get-application-warning-2",
								}, expectedErr)
						})

						It("return the same error and outputs the warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-application-warning-1"))
							Expect(testUI.Err).To(Say("get-application-warning-2"))
						})
					})

					Context("when GetApplicationTasks returns an error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("bananapants??")
							fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "some-app-guid"},
								v3action.Warnings{
									"get-application-warning-1",
									"get-application-warning-2",
								}, nil)
							fakeActor.GetApplicationTasksReturns(nil,
								v3action.Warnings{
									"get-tasks-warning-1",
									"get-tasks-warning-2",
								}, expectedErr)
						})

						It("returns the same error and outputs all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-application-warning-1"))
							Expect(testUI.Err).To(Say("get-application-warning-2"))
							Expect(testUI.Err).To(Say("get-tasks-warning-1"))
							Expect(testUI.Err).To(Say("get-tasks-warning-2"))
						})
					})
				})
			})
		})
	})
})
