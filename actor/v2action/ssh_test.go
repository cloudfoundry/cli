package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSH Actions", func() {
	var (
		actor         *Actor
		fakeConfig    *v2actionfakes.FakeConfig
		fakeUAAClient *v2actionfakes.FakeUAAClient
	)

	BeforeEach(func() {
		fakeConfig = new(v2actionfakes.FakeConfig)
		fakeUAAClient = new(v2actionfakes.FakeUAAClient)
		actor = NewActor(nil, fakeUAAClient, fakeConfig)
	})

	Describe("GetSSHPasscode", func() {
		var uaaAccessToken string

		BeforeEach(func() {
			uaaAccessToken = "4cc3sst0k3n"
			fakeConfig.AccessTokenReturns(uaaAccessToken)
			fakeConfig.SSHOAuthClientReturns("some-id")
		})

		Context("when no errors are encountered getting the ssh passcode", func() {
			var expectedCode string

			BeforeEach(func() {
				expectedCode = "s3curep4ss"
				fakeUAAClient.GetSSHPasscodeReturns(expectedCode, nil)
			})

			It("returns the ssh passcode", func() {
				code, err := actor.GetSSHPasscode()
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(Equal(expectedCode))
				Expect(fakeUAAClient.GetSSHPasscodeCallCount()).To(Equal(1))
				accessTokenArg, sshOAuthClientArg := fakeUAAClient.GetSSHPasscodeArgsForCall(0)
				Expect(accessTokenArg).To(Equal(uaaAccessToken))
				Expect(sshOAuthClientArg).To(Equal("some-id"))
			})
		})

		Context("when an error is encountered getting the ssh passcode", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("failed fetching code")
				fakeUAAClient.GetSSHPasscodeReturns("", expectedErr)
			})

			It("returns the error", func() {
				_, err := actor.GetSSHPasscode()
				Expect(err).To(MatchError(expectedErr))
				Expect(fakeUAAClient.GetSSHPasscodeCallCount()).To(Equal(1))
				accessTokenArg, sshOAuthClientArg := fakeUAAClient.GetSSHPasscodeArgsForCall(0)
				Expect(accessTokenArg).To(Equal(uaaAccessToken))
				Expect(sshOAuthClientArg).To(Equal("some-id"))
			})
		})
	})
})
