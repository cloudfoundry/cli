package helpers_test

import (
	"unicode/utf8"

	"code.cloudfoundry.org/diego-ssh/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.org/x/crypto/ssh"
)

const (
	TestPrivateKeyPem = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAx0y65jB977anY39jzB7AkojdAyqiADG4BTcXmKIy7w/GY/bi
Aq/AcO/SVAsq1iJ+SAmiQXt1K6kL8wUGtxlB1D+Ze0d0jw05Ep+O5rRF1dMDFUsA
0yrgfsUfryl7XOl9LmE1PKLinKExooZrfJTSqW1oRjHpIWZMtJj25glDtrfz+Wyd
6wPYH1CgOdQjSAORWYrhb9xYPzIxG5XMeMyO5xtL3sLEsdM1iROIl9rOu1qem1n+
Xs/z+03tpIoQc4gjVvbykYdZQ5mvCgbxtPRVy7EmYSbQflmMu3TWZmT20ZWDCHzl
3eDdFMOpqRaRFg/TjlNjH20TQbruKumk7ldogQIDAQABAoIBABujCEfjcZNMQOoL
QEuN+CZZ1EwcHVrpihsvCJah524/QcOa+LxmoskGeKQu6EHJhrl2nIl4FUd4qa+J
guThG7/TEfWGcyNjMgbjGW3kkcqU+Fh7jiG6UGdD7qDbn7/CoRlNYZSHAeW2dKuU
+FLOUGguQ8d4JFv9U6W3kIVVw44StVMkQwh0TB9kh7yzeHrpVddaMPzVZUmCWm2Y
NPN5EZq96DmmcEQC7Gktj7kPgC5UWcc8wF2Xy74sZb3RKOeyc5e7ddMDLbNI5STr
iRT3Fg+bhWQhhMUQfvD9KSh/9IK0OGu/3SSb9WeEzMUdh5mho1IsERugaXsVlne8
6JWW7gECgYEAzTTSJDRm8CiBSa4sn5KzLOHvn3YfSC91aERjrbZuVDdmVMJvhpLw
JW6/5zmz7X7Hr3mwHBSj+rS4/rIoEVvTjWrJm6GUSXXPwRwoJedbK67FU0MxBMzt
iqi+qBHdsKRhdrlM3W9RryGkcS1AkK+6B2Feu3GVGUQDz6G/yaTJ8UcCgYEA+KGh
D+PtdAd3s1sdAJlRuS4kCXCLbO/5EWfMMHVaewebpGs8bZnW4cpFaGR7zXd9Emkk
QuZWE7L44SNQrirECtGcu3zEKx1grYo+2jYoLYexiwOf6UEMWJEExLS475EDgmUJ
7Fy5tt2mwwV2GBXZfTHuQLOo9Zxjsf3NAKAZ+/cCgYBV+nKtrrMOnroE6BBUT7/4
5zViJ7jVouTbagQlrZEuggPDMbBOv1QVKwEG3Ztwv7Tk5eSO72sBSSVVucml9EaA
MyUDq0CZQt5oN+bucrA1bkXJLBbmvwIsHaW8f7fWIhmgB+WXxeOAsGTY8q/hr28P
VpG9kcp5ypCaN1hHIV9nUwKBgQDKcUBlYd8MJLBwV3XL8Qq7zzgEf6Dm+JZCd9Oo
eUVM+6rdO3ueei6e9kWBdJ/hcrNh9D5UQpw/ufAv0MN2rNenP3lwp2xK9sarRu9a
WdJpEB2d5TulfxOAYcQSLlyOo/LJj19/FxkYLm4ESUQY5GGMMMWf5Sljow0B9nef
VL0TjQKBgG9/w5XpX7K8nnUVGgYuEhbBj7lel2Ad7wjqwxuqDxi3jqVvuIR7VYeg
feuxbZkmphtEOKtaVDSWxGbNXbuN8H9eQqsGhK1Xcn/FxKVu7k+9GYyqeOwhjaqy
HbXzxBM4Ki0l1kaUjDVKjz3fsIq9Pl/lBoKYAmDvkK4xoxcs05ws
-----END RSA PRIVATE KEY-----`

	ExpectedMD5Fingerprint  = `24:2e:53:c3:72:4f:25:b8:72:29:2d:e3:56:63:4b:c8`
	ExpectedSHA1Fingerprint = `8b:d1:ce:b8:3a:f0:37:7f:56:9e:33:1a:72:4b:32:5a:bc:9d:3b:49`
)

var _ = Describe("Fingerprint", func() {
	var publicKey ssh.PublicKey
	var fingerprint string

	BeforeEach(func() {
		privateKey, err := ssh.ParsePrivateKey([]byte(TestPrivateKeyPem))
		Expect(err).NotTo(HaveOccurred())

		publicKey = privateKey.PublicKey()
	})

	Describe("MD5 Fingerprint", func() {
		BeforeEach(func() {
			fingerprint = helpers.MD5Fingerprint(publicKey)
		})

		It("should have the correct length", func() {
			Expect(utf8.RuneCountInString(fingerprint)).To(Equal(helpers.MD5_FINGERPRINT_LENGTH))
		})

		It("should match the expected fingerprint", func() {
			Expect(fingerprint).To(Equal(ExpectedMD5Fingerprint))
		})
	})

	Describe("SHA1 Fingerprint", func() {
		BeforeEach(func() {
			fingerprint = helpers.SHA1Fingerprint(publicKey)
		})

		It("should have the correct length", func() {
			Expect(utf8.RuneCountInString(fingerprint)).To(Equal(helpers.SHA1_FINGERPRINT_LENGTH))
		})

		It("should match the expected fingerprint", func() {
			Expect(fingerprint).To(Equal(ExpectedSHA1Fingerprint))
		})
	})
})
