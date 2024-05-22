package clissh_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var (
	TestHostKey    ssh.Signer
	TestPrivateKey ssh.Signer
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI SSH Suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(3 * time.Second)

	hostKeyBytes, err := os.ReadFile(filepath.Join("..", "..", "fixtures", "host-key"))
	Expect(err).NotTo(HaveOccurred())
	hostKey, err := ssh.ParsePrivateKey(hostKeyBytes)
	Expect(err).NotTo(HaveOccurred())

	privateKeyBytes, err := os.ReadFile(filepath.Join("..", "..", "fixtures", "private-key"))
	Expect(err).NotTo(HaveOccurred())
	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
	Expect(err).NotTo(HaveOccurred())

	TestHostKey = hostKey
	TestPrivateKey = privateKey
})

var _ = BeforeEach(func() {
	log.SetLevel(log.PanicLevel)
})
