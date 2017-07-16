package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Actions", func() {
	var (
		actor         *Actor
		fakeUAAClient *v2actionfakes.FakeUAAClient
		fakeConfig    *v2actionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeUAAClient = new(v2actionfakes.FakeUAAClient)
		actor = NewActor(nil, fakeUAAClient, nil)
		fakeConfig = new(v2actionfakes.FakeConfig)
	})

	Describe("Authenticate", func() {
		var actualErr error

		JustBeforeEach(func() {
			actualErr = actor.Authenticate(fakeConfig, "some-username", "some-password")
		})

		Context("when no API errors occur", func() {
			BeforeEach(func() {
				fakeUAAClient.AuthenticateReturns(
					"some-access-token",
					"some-refresh-token",
					nil,
				)
			})

			It("authenticates the user and returns access and refresh tokens", func() {
				Expect(actualErr).NotTo(HaveOccurred())

				Expect(fakeUAAClient.AuthenticateCallCount()).To(Equal(1))
				username, password := fakeUAAClient.AuthenticateArgsForCall(0)
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))

				Expect(fakeConfig.SetTokenInformationCallCount()).To(Equal(1))
				accessToken, refreshToken, sshOAuthClient := fakeConfig.SetTokenInformationArgsForCall(0)
				Expect(accessToken).To(Equal("bearer some-access-token"))
				Expect(refreshToken).To(Equal("some-refresh-token"))
				Expect(sshOAuthClient).To(BeEmpty())

				Expect(fakeConfig.UnsetOrganizationInformationCallCount()).To(Equal(1))
				Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(1))
			})
		})

		Context("when an API error occurs", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeUAAClient.AuthenticateReturns(
					"",
					"",
					expectedErr,
				)
			})

			It("returns the error", func() {
				Expect(actualErr).To(MatchError(expectedErr))

				Expect(fakeUAAClient.AuthenticateCallCount()).To(Equal(1))
				username, password := fakeUAAClient.AuthenticateArgsForCall(0)
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))

				Expect(fakeConfig.SetTokenInformationCallCount()).To(Equal(1))
				accessToken, refreshToken, sshOAuthClient := fakeConfig.SetTokenInformationArgsForCall(0)
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
				Expect(sshOAuthClient).To(BeEmpty())

				Expect(fakeConfig.UnsetOrganizationInformationCallCount()).To(Equal(1))
				Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(1))
			})
		})
	})
})
