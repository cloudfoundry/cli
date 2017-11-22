package keys_test

import (
	"crypto/x509"
	"encoding/pem"

	"code.cloudfoundry.org/diego-ssh/helpers"
	"code.cloudfoundry.org/diego-ssh/keys"
	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RSA", func() {
	var keyPair keys.KeyPair
	var bits int

	BeforeEach(func() {
		bits = 1024
	})

	JustBeforeEach(func() {
		var err error
		keyPair, err = keys.RSAKeyPairFactory.NewKeyPair(bits)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("PrivateKey", func() {
		It("returns the ssh private key associted with the public key", func() {
			Expect(keyPair.PrivateKey()).NotTo(BeNil())
			Expect(keyPair.PrivateKey().PublicKey()).To(Equal(keyPair.PublicKey()))
		})

		Context("when creating a 1024 bit key", func() {
			BeforeEach(func() {
				bits = 1024
			})

			It("the private key is 1024 bits", func() {
				block, _ := pem.Decode([]byte(keyPair.PEMEncodedPrivateKey()))
				key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
				Expect(err).NotTo(HaveOccurred())

				Expect(key.N.BitLen()).To(Equal(1024))
			})
		})

		Context("when creating a 2048 bit key", func() {
			BeforeEach(func() {
				bits = 2048
			})

			It("the private key is 2048 bits", func() {
				block, _ := pem.Decode([]byte(keyPair.PEMEncodedPrivateKey()))
				key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
				Expect(err).NotTo(HaveOccurred())

				Expect(key.N.BitLen()).To(Equal(2048))
			})
		})
	})

	Describe("PEMEncodedPrivateKey", func() {
		It("correctly represents the private key", func() {
			privateKey, err := ssh.ParsePrivateKey([]byte(keyPair.PEMEncodedPrivateKey()))
			Expect(err).NotTo(HaveOccurred())

			Expect(privateKey.PublicKey().Marshal()).To(Equal(keyPair.PublicKey().Marshal()))
		})
	})

	Describe("PublicKey", func() {
		It("equals the public key associated with the private key", func() {
			Expect(keyPair.PrivateKey().PublicKey().Marshal()).To(Equal(keyPair.PublicKey().Marshal()))
		})
	})

	Describe("Fingerprint", func() {
		It("equals the MD5 fingerprint of the public key", func() {
			expectedFingerprint := helpers.MD5Fingerprint(keyPair.PublicKey())

			Expect(keyPair.Fingerprint()).To(Equal(expectedFingerprint))
		})
	})

	Describe("AuthorizedKey", func() {
		It("equals the authorized key formatted public key", func() {
			expectedAuthorizedKey := string(ssh.MarshalAuthorizedKey(keyPair.PublicKey()))

			Expect(keyPair.AuthorizedKey()).To(Equal(expectedAuthorizedKey))
		})
	})
})
