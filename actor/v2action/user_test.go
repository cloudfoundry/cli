package v2action_test

import (
	"errors"
	"fmt"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Actions", func() {
	var (
		actor                     Actor
		fakeUAAClient             *v2actionfakes.FakeUAAClient
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeUAAClient = new(v2actionfakes.FakeUAAClient)
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, fakeUAAClient)
	})

	Describe("NewUser", func() {
		var (
			actualUser     User
			actualWarnings Warnings
			actualErr      error
		)

		JustBeforeEach(func() {
			actualUser, actualWarnings, actualErr = actor.NewUser("some-new-user", "some-password")
		})

		Context("when no API errors occur", func() {
			var createdUser ccv2.User

			BeforeEach(func() {
				createdUser = ccv2.User{
					GUID: "new-user-cc-guid",
				}
				fakeUAAClient.NewUserReturns(
					uaa.User{
						ID: "new-user-uaa-id",
					},
					nil,
				)
				fakeCloudControllerClient.NewUserReturns(
					createdUser,
					ccv2.Warnings{
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

				Expect(fakeUAAClient.NewUserCallCount()).To(Equal(1))
				username, password := fakeUAAClient.NewUserArgsForCall(0)
				Expect(username).To(Equal("some-new-user"))
				Expect(password).To(Equal("some-password"))

				Expect(fakeCloudControllerClient.NewUserCallCount()).To(Equal(1))
				uaaUserID := fakeCloudControllerClient.NewUserArgsForCall(0)
				Expect(uaaUserID).To(Equal("new-user-uaa-id"))
			})
		})

		Context("when a create user request to the UAA returns an error", func() {
			Context("when we get a 409 status code because the user already exists", func() {
				BeforeEach(func() {
					fakeUAAClient.NewUserReturns(
						uaa.User{},
						uaa.ConflictError{Message: "Username already in use: some-new-user"},
					)
				})

				It("does not return an error and returns a user already exists warning", func() {
					Expect(actualErr).NotTo(HaveOccurred())
					Expect(actualWarnings).To(ConsistOf(fmt.Sprintf("user some-new-user already exists")))
				})
			})

			Context("when the user does not exist", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = errors.New("UAA error")
					fakeUAAClient.NewUserReturns(uaa.User{}, returnedErr)
				})

				It("returns the error", func() {
					Expect(actualErr).To(MatchError(returnedErr))
				})
			})
		})

		Context("when the CC API returns an error", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("CC error")
				fakeUAAClient.NewUserReturns(
					uaa.User{
						ID: "new-user-uaa-id",
					},
					nil,
				)
				fakeCloudControllerClient.NewUserReturns(
					ccv2.User{},
					ccv2.Warnings{
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
})
