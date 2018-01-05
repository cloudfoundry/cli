package clissh_test

import (
	"io/ioutil"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"

	"testing"
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

	hostKeyBytes, err := ioutil.ReadFile(filepath.Join("..", "..", "fixtures", "host-key"))
	Expect(err).NotTo(HaveOccurred())
	hostKey, err := ssh.ParsePrivateKey(hostKeyBytes)
	Expect(err).NotTo(HaveOccurred())

	privateKeyBytes, err := ioutil.ReadFile(filepath.Join("..", "..", "fixtures", "private-key"))
	Expect(err).NotTo(HaveOccurred())
	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
	Expect(err).NotTo(HaveOccurred())

	TestHostKey = hostKey
	TestPrivateKey = privateKey
})
