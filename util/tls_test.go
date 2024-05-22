package util_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"

	. "code.cloudfoundry.org/cli/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TLS", func() {
	Describe("NewTLSConfig", func() {
		var (
			tlsConfig *tls.Config
		)

		BeforeEach(func() {
			tlsConfig = NewTLSConfig(nil, false)
		})

		It("sets minimum version of TLS to 1.2", func() {
			Expect(tlsConfig.MinVersion).To(BeEquivalentTo(tls.VersionTLS12))
		})

		It("sets maximum version of TLS to 1.3", func() {
			Expect(tlsConfig.MaxVersion).To(BeEquivalentTo(tls.VersionTLS13))
		})

		It("does not have any trusted CAs", func() {
			Expect(tlsConfig.RootCAs).To(BeNil())
		})

		It("verifies certificates", func() {
			Expect(tlsConfig.InsecureSkipVerify).To(BeFalse())
		})

		When("trusted certificates are provided", func() {
			var (
				certPEM = []byte(`-----BEGIN CERTIFICATE-----
MIICNTCCAZ6gAwIBAgIQeXuK80BdTIYBjxChLKAvRzANBgkqhkiG9w0BAQsFADAj
MSEwHwYDVQQKExhDbG91ZCBGb3VuZHJ5IEZvdW5kYXRpb24wIBcNNzAwMTAxMDAw
MDAwWhgPMjA4NDAxMjkxNjAwMDBaMCMxITAfBgNVBAoTGENsb3VkIEZvdW5kcnkg
Rm91bmRhdGlvbjCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEAwDVHeEOGoTty
jdKHS4EKLnCataMnUsa+uRuLi5WgPUI/R6ufb63Yex8u76It3YMjiDRqgI8g/fyO
vScso8mLmjFMdbNcMRAKsqARksKSwAasupmRVUlF3F8+8bgT1c5P82wD8nSb7zzy
KC2VDZtc1kwsJDCQVm47Tkp+nP5Z73UCAwEAAaNoMGYwDgYDVR0PAQH/BAQDAgKk
MBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQFMAMBAf8wLgYDVR0RBCcw
JYILZXhhbXBsZS5jb22HBH8AAAGHEAAAAAAAAAAAAAAAAAAAAAEwDQYJKoZIhvcN
AQELBQADgYEAeHp7bEIML8KL7UpwIhFxrwXYVYbIxGsn/0ret0DKwmyqcNVoFpg+
FNGTopDe0V2O8/0ZdxQuiYGoARRe266AuNOhAYBeyIZpnf9Ypt78V+21YACHQ4YL
RVNEplh5ZEYbbWclddUBf46JPRU/eEu4JMqOJOykTdwbByFa3909Bzs=
-----END CERTIFICATE-----`)
				cert *x509.Certificate
			)

			BeforeEach(func() {
				var err error

				block, _ := pem.Decode(certPEM)
				Expect(block).ToNot(BeNil())
				cert, err = x509.ParseCertificate(block.Bytes)
				Expect(err).ToNot(HaveOccurred())
				tlsConfig = NewTLSConfig([]*x509.Certificate{cert}, false)
			})

			It("adds them to the trusted CAs", func() {
				Expect(tlsConfig.RootCAs.Subjects()).To(ContainElement(ContainSubstring("Cloud Foundry")))
			})
		})

		When("skipSSLValidation is true", func() {
			BeforeEach(func() {
				tlsConfig = NewTLSConfig(nil, true)
			})
			It("does not verify certificates", func() {
				Expect(tlsConfig.InsecureSkipVerify).To(BeTrue())
			})
		})

	})

})
