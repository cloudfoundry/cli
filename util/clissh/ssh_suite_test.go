package clissh_test

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	// Suppress log output during tests. This equates with setting to panic-level severity in older logging libraries
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

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
