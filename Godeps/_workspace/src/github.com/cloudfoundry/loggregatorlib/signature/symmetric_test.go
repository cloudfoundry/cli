package signature

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Encryption", func() {
	var message = []byte("Super secret message that no one should read")

	It("encrypts and decrypts with a 'long' key", func() {
		key := "aaaaaaaaaaaaaaaa"
		encrypted, err := Encrypt(key, message)
		Expect(err).NotTo(HaveOccurred())

		decrypted, err := Decrypt(key, encrypted)
		Expect(err).NotTo(HaveOccurred())

		Expect(decrypted).To(Equal(message))
		Expect(encrypted).NotTo(Equal(message))
	})

	It("encrypts and decrypts with a 'short' key", func() {
		key := "short key"
		encrypted, err := Encrypt(key, message)
		Expect(err).NotTo(HaveOccurred())

		decrypted, err := Decrypt(key, encrypted)
		Expect(err).NotTo(HaveOccurred())

		Expect(decrypted).To(Equal(message))
		Expect(encrypted).NotTo(Equal(message))
	})

	It("fails to decrypt with the wrong key", func() {
		key := "short key"
		encrypted, err := Encrypt(key, message)
		Expect(err).NotTo(HaveOccurred())

		_, err = Decrypt("wrong key", encrypted)
		Expect(err).To(HaveOccurred())
	})

	It("is non-deterministic", func() {
		key := "aaaaaaaaaaaaaaaa"
		encrypted1, err := Encrypt(key, message)
		Expect(err).NotTo(HaveOccurred())

		encrypted2, err := Encrypt(key, message)
		Expect(err).NotTo(HaveOccurred())

		Expect(encrypted1).NotTo(Equal(encrypted2))
	})

	It("correctly computes a digest", func() {
		Expect(DigestBytes([]byte("some-key"))).To(Equal([]byte{0x68, 0x2f, 0x66, 0x97, 0xfa, 0x93, 0xec, 0xa6, 0xc8, 0x1, 0xa2, 0x32, 0x51, 0x9a, 0x9, 0xe3, 0xfe, 0xc, 0x5c, 0x33, 0x94, 0x65, 0xee, 0x53, 0xc3, 0xf9, 0xed, 0xf9, 0x2f, 0xd0, 0x1f, 0x35}))
	})

	It("deterministically computes the actual encryption key", func() {
		key := "12345"
		Expect(getEncryptionKey(key)).To(Equal([]byte{0x59, 0x94, 0x47, 0x1a, 0xbb, 0x1, 0x11, 0x2a, 0xfc, 0xc1, 0x81, 0x59, 0xf6, 0xcc, 0x74, 0xb4}))
	})
})
