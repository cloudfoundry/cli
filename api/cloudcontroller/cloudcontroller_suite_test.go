package cloudcontroller_test

import (
	"bytes"
	"log"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

func TestCloudcontroller(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cloud Controller Suite")
}

var server *Server

var _ = SynchronizedBeforeSuite(func() []byte {
	return []byte{}
}, func(data []byte) {
	server = NewTLSServer()

	// Suppresses ginkgo server logs
	server.HTTPTestServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
})

var _ = SynchronizedAfterSuite(func() {
	server.Close()
}, func() {})

var _ = BeforeEach(func() {
	server.Reset()
})
