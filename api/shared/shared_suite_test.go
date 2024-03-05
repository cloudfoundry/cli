package shared_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestShared(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shared Wrapper Suite")
}

var (
	keyPair *rsa.PrivateKey
)

var _ = BeforeEach(func() {
	var err error
	keyPair, err = rsa.GenerateKey(rand.Reader, 2048)
	Expect(err).NotTo(HaveOccurred())
})
