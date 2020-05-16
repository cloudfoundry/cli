package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/resources"
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
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, fakeUAAClient, nil, nil)
	})

	Describe("CreateUser", func() {
		var (
			actualUser     resources.User
			actualWarnings Warnings
			actualErr      error
		)

		JustBeforeEach(func() {
			actualUser, actualWarnings, actualErr = actor.CreateUser("some-new-user", "some-password", "some-origin")
		})

		When("no API errors occur", func() {
			var createdUser resources.User

			BeforeEach(func() {
				createdUser = resources.User{
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

				Expect(actualUser).To(Equal(resources.User(createdUser)))
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
					resources.User{},
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
			actualUser resources.User
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
					Expect(actualUser).To(Equal(resources.User{GUID: "user-id"}))

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
					Expect(actualUser).To(Equal(resources.User{}))
					Expect(actualErr).To(Equal(actionerror.UserNotFoundError{
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
					Expect(actualUser).To(Equal(resources.User{}))
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
				Expect(actualUser).To(Equal(resources.User{}))
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
					"",
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
						"",
						ccv3.Warnings{},
						returnedErr,
					)
				})

				It("does Not return the error", func() {
					Expect(actualErr).To(BeNil())
				})
			})
		})

		When("the CC API returns a job", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteUserReturns(
					"some-url",
					ccv3.Warnings{"warning-5", "warning-6"},
					nil,
				)
			})

			When("polling errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("Never expected, by anyone")
					fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"warning-7", "warning-8"}, expectedErr)
				})

				It("returns the error", func() {
					Expect(actualErr).To(Equal(expectedErr))
					Expect(actualWarnings).To(ConsistOf("warning-5", "warning-6", "warning-7", "warning-8"))
				})
			})

			When("the job is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"warning-7", "warning-8"}, nil)
				})

				It("returns warnings and no error", func() {
					Expect(actualErr).ToNot(HaveOccurred())
					Expect(actualWarnings).To(ConsistOf("warning-5", "warning-6", "warning-7", "warning-8"))

					Expect(fakeCloudControllerClient.DeleteUserCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.DeleteUserArgsForCall(0)).To(Equal("some-user-guid"))

					Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.PollJobArgsForCall(0)).To(Equal(ccv3.JobURL("some-url")))
				})
			})
		})

	})

	Describe("UpdateUserPassword", func() {
		var (
			userGUID    = "some-user"
			oldPassword = "old1"
			newPassword = "new1"
			actualErr   error
		)

		JustBeforeEach(func() {
			actualErr = actor.UpdateUserPassword(userGUID, oldPassword, newPassword)
		})

		When("no API errors occur", func() {
			BeforeEach(func() {
				fakeUAAClient.UpdatePasswordReturns(nil)
			})

			It("does not return an error", func() {
				Expect(actualErr).NotTo(HaveOccurred())

				Expect(fakeUAAClient.UpdatePasswordCallCount()).To(Equal(1))
				givenUserGUID, givenOldPassword, givenNewPassword := fakeUAAClient.UpdatePasswordArgsForCall(0)
				Expect(givenUserGUID).To(Equal(userGUID))
				Expect(givenOldPassword).To(Equal(oldPassword))
				Expect(givenNewPassword).To(Equal(newPassword))
			})
		})

		When("the UAA API returns an error", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("some UAA error")
				fakeUAAClient.UpdatePasswordReturns(returnedErr)
			})

			It("returns the same error", func() {
				Expect(actualErr).To(MatchError(returnedErr))

				Expect(fakeUAAClient.UpdatePasswordCallCount()).To(Equal(1))
				givenUserGUID, givenOldPassword, givenNewPassword := fakeUAAClient.UpdatePasswordArgsForCall(0)
				Expect(givenUserGUID).To(Equal(userGUID))
				Expect(givenOldPassword).To(Equal(oldPassword))
				Expect(givenNewPassword).To(Equal(newPassword))
			})
		})
	})

	Describe("SortUsers", func() {
		var (
			users []resources.User
		)
		When("The PresentationNames are different", func() {
			BeforeEach(func() {
				users = []resources.User{
					{PresentationName: "c", Origin: "uaa"},
					{PresentationName: "a", Origin: "uaa"},
					{PresentationName: "b", Origin: "ldap"},
				}
			})
			It("sorts by PresentationName", func() {
				SortUsers(users)
				Expect(users[0].PresentationName).To(Equal("a"))
				Expect(users[1].PresentationName).To(Equal("b"))
				Expect(users[2].PresentationName).To(Equal("c"))
			})
		})
		When("The PresentationNames are identical", func() {
			BeforeEach(func() {
				users = []resources.User{
					{PresentationName: "a", Origin: "cc"},
					{PresentationName: "a", Origin: "aa"},
					{PresentationName: "a", Origin: "bb"},
					{PresentationName: "a", Origin: "uaa"},
					{PresentationName: "a", Origin: "zz"},
					{PresentationName: "a", Origin: ""},
				}
			})
			It("sorts by PresentationName, uaa first, clients (origin == '') last and alphabetically otherwise", func() {
				SortUsers(users)
				Expect(users[0].Origin).To(Equal("uaa"))
				Expect(users[1].Origin).To(Equal("aa"))
				Expect(users[2].Origin).To(Equal("bb"))
				Expect(users[3].Origin).To(Equal("cc"))
				Expect(users[4].Origin).To(Equal("zz"))
				Expect(users[5].Origin).To(Equal(""))
			})
		})
	})

	Describe("GetHumanReadableOrigin", func() {
		var user resources.User
		When("The user has an origin", func() {
			BeforeEach(func() {
				user = resources.User{Origin: "some-origin"}
			})
			It("displays the origin", func() {
				Expect(GetHumanReadableOrigin(user)).To(Equal("some-origin"))
			})
		})
		When("The user has an empty origin", func() {
			BeforeEach(func() {
				user = resources.User{Origin: ""}
			})
			It("displays 'client'", func() {
				Expect(GetHumanReadableOrigin(user)).To(Equal("client"))
			})
		})
	})
})
