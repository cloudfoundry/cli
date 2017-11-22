package proxy_test

import (
	"code.cloudfoundry.org/diego-ssh/keys"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"

	"testing"
)

var (
	TestHostKey ssh.Signer

	TestPrivatePem          string
	TestPublicAuthorizedKey string
)

func TestProxy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Proxy Suite")
}

var _ = BeforeSuite(func() {
	hostKey, err := keys.RSAKeyPairFactory.NewKeyPair(1024)
	Expect(err).NotTo(HaveOccurred())

	privateKey, err := keys.RSAKeyPairFactory.NewKeyPair(1024)
	Expect(err).NotTo(HaveOccurred())

	TestHostKey = hostKey.PrivateKey()
	TestPrivatePem = privateKey.PEMEncodedPrivateKey()
	TestPublicAuthorizedKey = privateKey.AuthorizedKey()
})
