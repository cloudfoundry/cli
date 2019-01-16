package buildpacks_test

import (
	"bytes"
	"log"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

func TestBuildpacks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Buildpacks Suite")
}

var server *Server

var _ = BeforeEach(func() {
	server = NewTLSServer()

	// Suppresses ginkgo server logs
	server.HTTPTestServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
})

var _ = AfterEach(func() {
	server.Close()
})
