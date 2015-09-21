package cmd_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/diego-ssh/helpers"
	"github.com/cloudfoundry-incubator/diego-ssh/keys"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"

	"testing"
)

var (
	TestHostKey            ssh.Signer
	TestHostKeyFingerprint string
	TestPrivateKey         ssh.Signer
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	hostKey, err := keys.RSAKeyPairFactory.NewKeyPair(1024)
	Expect(err).NotTo(HaveOccurred())

	privateKey, err := keys.RSAKeyPairFactory.NewKeyPair(1024)
	Expect(err).NotTo(HaveOccurred())

	payload, err := json.Marshal(map[string]string{
		"host-key":    hostKey.PEMEncodedPrivateKey(),
		"private-key": privateKey.PEMEncodedPrivateKey(),
	})

	Expect(err).NotTo(HaveOccurred())

	return payload

}, func(payload []byte) {
	context := map[string]string{}

	err := json.Unmarshal(payload, &context)
	Expect(err).NotTo(HaveOccurred())

	hostKeyPem := context["host-key"]
	hostKey, err := ssh.ParsePrivateKey([]byte(hostKeyPem))
	Expect(err).NotTo(HaveOccurred())
	TestHostKey = hostKey

	privateKeyPem := context["private-key"]
	privateKey, err := ssh.ParsePrivateKey([]byte(privateKeyPem))
	Expect(err).NotTo(HaveOccurred())
	TestPrivateKey = privateKey

	TestHostKeyFingerprint = helpers.SHA1Fingerprint(TestHostKey.PublicKey())
})
