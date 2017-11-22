package authenticators_test

import (
	"errors"
	"regexp"

	"code.cloudfoundry.org/diego-ssh/authenticators"
	"code.cloudfoundry.org/diego-ssh/authenticators/fake_authenticators"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_ssh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
)

var _ = Describe("CompositeAuthenticator", func() {
	Describe("Authenticate", func() {
		var (
			authenticator *authenticators.CompositeAuthenticator
			authens       []authenticators.PasswordAuthenticator
			metadata      *fake_ssh.FakeConnMetadata
			password      []byte
		)

		BeforeEach(func() {
			authens = []authenticators.PasswordAuthenticator{}
			metadata = &fake_ssh.FakeConnMetadata{}
			password = []byte{}
		})

		JustBeforeEach(func() {
			authenticator = authenticators.NewCompositeAuthenticator(authens...)
		})

		Context("when no authenticators are specified", func() {
			It("fails to authenticate", func() {
				_, err := authenticator.Authenticate(metadata, password)
				Expect(err).To(Equal(authenticators.InvalidCredentialsErr))
			})
		})

		Context("when one or more authenticators are specified", func() {
			var (
				authenticatorOne *fake_authenticators.FakePasswordAuthenticator
				authenticatorTwo *fake_authenticators.FakePasswordAuthenticator
			)

			BeforeEach(func() {
				authenticatorOne = &fake_authenticators.FakePasswordAuthenticator{}
				authenticatorOne.UserRegexpReturns(regexp.MustCompile("one:.*"))

				authenticatorTwo = &fake_authenticators.FakePasswordAuthenticator{}
				authenticatorTwo.UserRegexpReturns(regexp.MustCompile("two:.*"))

				authens = []authenticators.PasswordAuthenticator{
					authenticatorOne,
					authenticatorTwo,
				}
			})

			Context("and the users realm matches the first authenticator", func() {
				BeforeEach(func() {
					metadata.UserReturns("one:garbage")
				})

				Context("and the authenticator successfully authenticates", func() {
					var permissions *ssh.Permissions

					BeforeEach(func() {
						permissions = &ssh.Permissions{}
						authenticatorOne.AuthenticateReturns(permissions, nil)
					})

					It("succeeds to authenticate", func() {
						perms, err := authenticator.Authenticate(metadata, password)

						Expect(err).NotTo(HaveOccurred())
						Expect(perms).To(Equal(permissions))
					})

					It("should provide the metadata to the authenticator", func() {
						_, err := authenticator.Authenticate(metadata, password)
						Expect(err).NotTo(HaveOccurred())
						m, p := authenticatorOne.AuthenticateArgsForCall(0)

						Expect(m).To(Equal(metadata))
						Expect(p).To(Equal(password))
					})
				})

				Context("and the authenticator fails to authenticate", func() {
					BeforeEach(func() {
						authenticatorOne.AuthenticateReturns(nil, errors.New("boom"))
					})

					It("fails to authenticate", func() {
						_, err := authenticator.Authenticate(metadata, password)
						Expect(err).To(MatchError("boom"))
					})
				})

				It("does not attempt to authenticate with any other authenticators", func() {
					authenticator.Authenticate(metadata, password)
					Expect(authenticatorTwo.AuthenticateCallCount()).To(Equal(0))
				})
			})

			Context("and the user realm is not valid", func() {
				BeforeEach(func() {
					metadata.UserReturns("one")
				})

				It("fails to authenticate", func() {
					_, err := authenticator.Authenticate(metadata, password)

					Expect(err).To(Equal(authenticators.InvalidCredentialsErr))
					Expect(authenticatorOne.AuthenticateCallCount()).To(Equal(0))
					Expect(authenticatorTwo.AuthenticateCallCount()).To(Equal(0))
				})
			})

			Context("and the user realm does not match any authenticators", func() {
				BeforeEach(func() {
					metadata.UserReturns("jim:")
				})

				It("fails to authenticate", func() {
					_, err := authenticator.Authenticate(metadata, password)

					Expect(err).To(Equal(authenticators.InvalidCredentialsErr))
					Expect(authenticatorOne.AuthenticateCallCount()).To(Equal(0))
					Expect(authenticatorTwo.AuthenticateCallCount()).To(Equal(0))
				})
			})
		})
	})
})
