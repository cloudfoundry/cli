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

var _ = Describe("RunTask Command", func() {
	var (
		cmd        v3.RunTaskCommand
		testUI     *ui.UI
		fakeActor  *v3fakes.FakeRunTaskActor
		fakeConfig *commandfakes.FakeConfig
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeActor = new(v3fakes.FakeRunTaskActor)
		fakeActor.CloudControllerAPIVersionReturns("3.0.0")
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.ExperimentalReturns(true)

		cmd = v3.RunTaskCommand{
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
				expectedErr = errors.New("got bananapants??")
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
					cmd.RequiredArgs.Command = "fake command"

					fakeActor.GetApplicationByNameAndSpaceReturns(
						v3action.Application{GUID: "some-app-guid"},
						v3action.Warnings{
							"get-application-warning-1",
							"get-application-warning-2",
						}, nil)
					fakeActor.RunTaskReturns(v3action.Task{SequenceID: 3},
						v3action.Warnings{
							"get-application-warning-3",
						}, nil)
				})

				Context("when the task name is not provided", func() {
					It("runs a new task and outputs all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
						appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
						Expect(appName).To(Equal("some-app-name"))
						Expect(spaceGUID).To(Equal("some-space-guid"))

						Expect(fakeActor.RunTaskCallCount()).To(Equal(1))
						appGUID, command, name := fakeActor.RunTaskArgsForCall(0)
						Expect(appGUID).To(Equal("some-app-guid"))
						Expect(command).To(Equal("fake command"))
						Expect(name).To(Equal(""))

						Expect(testUI.Out).To(Say(`Creating task for app some-app-name in org some-org / space some-space as some-user...
OK

Task 3 has been submitted successfully for execution.`))
						Expect(testUI.Err).To(Say(`get-application-warning-1
get-application-warning-2
get-application-warning-3`))
					})
				})

				Context("when the task name is provided", func() {
					BeforeEach(func() {
						cmd.Name = "some-task-name"
					})

					It("runs a new task and outputs all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
						appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
						Expect(appName).To(Equal("some-app-name"))
						Expect(spaceGUID).To(Equal("some-space-guid"))

						Expect(fakeActor.RunTaskCallCount()).To(Equal(1))
						appGUID, command, name := fakeActor.RunTaskArgsForCall(0)
						Expect(appGUID).To(Equal("some-app-guid"))
						Expect(command).To(Equal("fake command"))
						Expect(name).To(Equal("some-task-name"))

						Expect(testUI.Out).To(Say(`Creating task for app some-app-name in org some-org / space some-space as some-user...
OK

Task 3 has been submitted successfully for execution.`,
						))
						Expect(testUI.Err).To(Say(`get-application-warning-1
get-application-warning-2
get-application-warning-3`))
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

					Context("when RunTask returns a translatable error", func() {
						var returnedErr error

						BeforeEach(func() {
							returnedErr = cloudcontroller.UnverifiedServerError{URL: "some-url"}
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.RunTaskReturns(
								v3action.Task{},
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
							expectedErr = errors.New("got bananapants??")
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

					Context("when RunTask returns an error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("got bananapants??")
							fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "some-app-guid"},
								v3action.Warnings{
									"get-application-warning-1",
									"get-application-warning-2",
								}, nil)
							fakeActor.RunTaskReturns(v3action.Task{},
								v3action.Warnings{
									"run-task-warning-1",
									"run-task-warning-2",
								}, expectedErr)
						})

						It("returns the same error and outputs all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-application-warning-1"))
							Expect(testUI.Err).To(Say("get-application-warning-2"))
							Expect(testUI.Err).To(Say("run-task-warning-1"))
							Expect(testUI.Err).To(Say("run-task-warning-2"))
						})
					})
				})
			})
		})
	})
})
