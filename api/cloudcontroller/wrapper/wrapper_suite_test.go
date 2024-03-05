package wrapper_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

func TestCloudcontroller(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cloud Controller Wrapper Suite")
}

var (
	server  *Server
	keyPair *rsa.PrivateKey
)

var _ = SynchronizedBeforeSuite(func() []byte {
	return []byte{}
}, func(data []byte) {
	server = NewTLSServer()

	// Suppresses ginkgo server logs
	server.HTTPTestServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)

	var err error
	keyPair, err = rsa.GenerateKey(rand.Reader, 2048)
	Expect(err).NotTo(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {
	server.Close()
}, func() {})

var _ = BeforeEach(func() {
	server.Reset()
})
