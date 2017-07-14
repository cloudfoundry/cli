package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Actions", func() {
	var (
		actor                     *Actor
		fakeUAAClient             *v2actionfakes.FakeUAAClient
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeUAAClient = new(v2actionfakes.FakeUAAClient)
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, fakeUAAClient, nil)
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

		Context("when no API errors occur", func() {
			var createdUser ccv2.User

			BeforeEach(func() {
				createdUser = ccv2.User{
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

		Context("when a create user request to the UAA returns an error", func() {
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

		Context("when the CC API returns an error", func() {
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
