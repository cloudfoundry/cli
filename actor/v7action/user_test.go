package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Actions", func() {
	var (
		actor                     *Actor
		fakeUAAClient             *v7actionfakes.FakeUAAClient
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeUAAClient = new(v7actionfakes.FakeUAAClient)
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeConfig := new(v7actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, fakeUAAClient, nil)
	})

	Describe("CreateUser", func() {
		var (
			actualUser     User
			actualWarnings Warnings
			actualErr      error
		)

		JustBeforeEach(func() {
			actualUser, actualWarnings, actualErr = actor.CreateUser("some-new-user", "some-password", "some-origin")
		})

		When("no API errors occur", func() {
			var createdUser ccv3.User

			BeforeEach(func() {
				createdUser = ccv3.User{
					GUID: "new-user-cc-guid",
				}
				fakeUAAClient.CreateUserReturns(
					uaa.User{
						ID: "new-user-uaa-id",
					},
					nil,
				)
				fakeCloudControllerClient.CreateUserReturns(
					createdUser,
					ccv3.Warnings{
						"warning-1",
						"warning-2",
					},
					nil,
				)
			})

			It("creates a new user and returns all warnings", func() {
				Expect(actualErr).NotTo(HaveOccurred())

				Expect(actualUser).To(Equal(User(createdUser)))
				Expect(actualWarnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeUAAClient.CreateUserCallCount()).To(Equal(1))
				username, password, origin := fakeUAAClient.CreateUserArgsForCall(0)
				Expect(username).To(Equal("some-new-user"))
				Expect(password).To(Equal("some-password"))
				Expect(origin).To(Equal("some-origin"))

				Expect(fakeCloudControllerClient.CreateUserCallCount()).To(Equal(1))
				uaaUserID := fakeCloudControllerClient.CreateUserArgsForCall(0)
				Expect(uaaUserID).To(Equal("new-user-uaa-id"))
			})
		})

		When("the UAA API returns an error", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("some UAA error")
				fakeUAAClient.CreateUserReturns(
					uaa.User{
						ID: "new-user-uaa-id",
					},
					returnedErr,
				)
			})

			It("returns the same error", func() {
				Expect(actualErr).To(MatchError(returnedErr))
			})
		})

		When("the CC API returns an error", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("CC error")
				fakeUAAClient.CreateUserReturns(
					uaa.User{
						ID: "new-user-uaa-id",
					},
					nil,
				)
				fakeCloudControllerClient.CreateUserReturns(
					ccv3.User{},
					ccv3.Warnings{
						"warning-1",
						"warning-2",
					},
					returnedErr,
				)
			})

			It("returns the same error and all warnings", func() {
				Expect(actualErr).To(MatchError(returnedErr))
				Expect(actualWarnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetUser", func() {
		var (
			actualUser User
			actualErr  error
		)

		JustBeforeEach(func() {
			actualUser, actualErr = actor.GetUser("some-user", "some-origin")
		})

		When("when the API returns a success response", func() {
			When("the API returns one user", func() {
				BeforeEach(func() {

					fakeUAAClient.ListUsersReturns(
						[]uaa.User{
							{ID: "user-id"},
						},
						nil,
					)
				})

				It("returns the single user", func() {
					Expect(actualErr).NotTo(HaveOccurred())
					Expect(actualUser).To(Equal(User{GUID: "user-id"}))

					Expect(fakeUAAClient.ListUsersCallCount()).To(Equal(1))
					username, origin := fakeUAAClient.ListUsersArgsForCall(0)
					Expect(username).To(Equal("some-user"))
					Expect(origin).To(Equal("some-origin"))
				})
			})

			When("the API returns no user", func() {
				BeforeEach(func() {
					fakeUAAClient.ListUsersReturns(
						[]uaa.User{},
						nil,
					)
				})

				It("returns an error indicating user was not found in UAA", func() {
					Expect(actualUser).To(Equal(User{}))
					Expect(actualErr).To(Equal(actionerror.UAAUserNotFoundError{
						Username: "some-user",
						Origin:   "some-origin",
					}))
					Expect(fakeUAAClient.ListUsersCallCount()).To(Equal(1))
				})
			})

			When("the API returns multiple users", func() {
				BeforeEach(func() {
					fakeUAAClient.ListUsersReturns(
						[]uaa.User{
							{ID: "user-guid-1", Origin: "uaa"},
							{ID: "user-guid-2", Origin: "ldap"},
						},
						nil,
					)
				})

				It("returns an error indicating user was not found in UAA", func() {
					Expect(actualUser).To(Equal(User{}))
					Expect(actualErr).To(Equal(actionerror.MultipleUAAUsersFoundError{
						Username: "some-user",
						Origins:  []string{"uaa", "ldap"},
					}))
					Expect(fakeUAAClient.ListUsersCallCount()).To(Equal(1))
				})
			})
		})

		When("the API returns an error", func() {
			var apiError error

			BeforeEach(func() {
				apiError = errors.New("uaa-api-get-users-error")
				fakeUAAClient.ListUsersReturns(
					nil,
					apiError,
				)
			})

			It("returns error coming from the API", func() {
				Expect(actualUser).To(Equal(User{}))
				Expect(actualErr).To(MatchError("uaa-api-get-users-error"))

				Expect(fakeUAAClient.ListUsersCallCount()).To(Equal(1))
			})
		})
	})

	Describe("DeleteUser", func() {
		var (
			actualWarnings Warnings
			actualErr      error
		)

		JustBeforeEach(func() {
			actualWarnings, actualErr = actor.DeleteUser("some-user-guid")
		})

		When("no API errors occur", func() {
			BeforeEach(func() {
				fakeUAAClient.DeleteUserReturns(
					uaa.User{
						ID: "some-user-guid",
					},
					nil,
				)
				fakeCloudControllerClient.DeleteUserReturns(
					ccv3.Warnings{
						"warning-1",
						"warning-2",
					},
					nil,
				)
			})

			It("Deletes user and returns all warnings", func() {
				Expect(actualErr).NotTo(HaveOccurred())

				Expect(actualWarnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeUAAClient.DeleteUserCallCount()).To(Equal(1))
				userGuid := fakeUAAClient.DeleteUserArgsForCall(0)
				Expect(userGuid).To(Equal("some-user-guid"))

				Expect(fakeCloudControllerClient.DeleteUserCallCount()).To(Equal(1))
				uaaUserID := fakeCloudControllerClient.DeleteUserArgsForCall(0)
				Expect(uaaUserID).To(Equal("some-user-guid"))
			})
		})

		When("the UAA API returns an error", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("some UAA error")
				fakeUAAClient.DeleteUserReturns(
					uaa.User{},
					returnedErr,
				)
			})

			It("returns the same error", func() {
				Expect(actualErr).To(MatchError(returnedErr))
			})
		})

		When("the CC API returns an error", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("some CC error")
				fakeUAAClient.DeleteUserReturns(
					uaa.User{},
					nil,
				)
				fakeCloudControllerClient.DeleteUserReturns(
					ccv3.Warnings{},
					returnedErr,
				)
			})

			It("returns the same error", func() {
				Expect(actualErr).To(MatchError(returnedErr))
			})

			When("the cloud controller found no user", func() {
				BeforeEach(func() {
					returnedErr = ccerror.ResourceNotFoundError{}
					fakeUAAClient.DeleteUserReturns(
						uaa.User{},
						nil,
					)
					fakeCloudControllerClient.DeleteUserReturns(
						ccv3.Warnings{},
						returnedErr,
					)
				})

				It("does Not return the error", func() {
					Expect(actualErr).To(BeNil())
				})
			})
		})
	})
})
