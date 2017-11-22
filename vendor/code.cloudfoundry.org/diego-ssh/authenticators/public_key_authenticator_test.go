package authenticators_test

import (
	"code.cloudfoundry.org/diego-ssh/authenticators"
	"code.cloudfoundry.org/diego-ssh/keys"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_ssh"
	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PublicKeyAuthenticator", func() {
	var (
		publicKey ssh.PublicKey

		authenticator authenticators.PublicKeyAuthenticator

		metadata  *fake_ssh.FakeConnMetadata
		clientKey ssh.PublicKey

		permissions *ssh.Permissions
		authnError  error
	)

	BeforeEach(func() {
		keyPair, err := keys.RSAKeyPairFactory.NewKeyPair(1024)
		Expect(err).NotTo(HaveOccurred())

		publicKey = keyPair.PublicKey()

		authenticator = authenticators.NewPublicKeyAuthenticator(publicKey)

		metadata = &fake_ssh.FakeConnMetadata{}
		clientKey = publicKey
	})

	JustBeforeEach(func() {
		permissions, authnError = authenticator.Authenticate(metadata, clientKey)
	})

	It("creates an authenticator", func() {
		Expect(authenticator).NotTo(BeNil())
		Expect(authenticator.PublicKey()).To(Equal(publicKey))
	})

	Describe("Authenticate", func() {
		BeforeEach(func() {
			clientKey = publicKey
		})

		Context("when the public key matches", func() {
			It("does not return an error", func() {
				Expect(authnError).NotTo(HaveOccurred())
				Expect(permissions).NotTo(BeNil())
			})
		})

		Context("when the public key does not match", func() {
			BeforeEach(func() {
				fakeKey := &fake_ssh.FakePublicKey{}
				fakeKey.MarshalReturns([]byte("go-away-alice"))
				clientKey = fakeKey
			})

			It("fails the authentication", func() {
				Expect(authnError).To(HaveOccurred())
				Expect(permissions).To(BeNil())
			})
		})
	})
})
