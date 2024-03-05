package ccv3_test

import (
	"bytes"
	"log"
	"testing"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

func TestCcv3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cloud Controller V3 Suite")
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

func NewFakeRequesterTestClient(requester Requester) (*Client, *ccv3fakes.FakeClock) {
	var client *Client
	fakeClock := new(ccv3fakes.FakeClock)

	client = TestClient(
		Config{AppName: "CF CLI API V3 Test", AppVersion: "Unknown"},
		fakeClock,
		requester,
	)

	return client, fakeClock
}

func NewTestClient(config ...Config) (*Client, *ccv3fakes.FakeClock) {
	var client *Client
	fakeClock := new(ccv3fakes.FakeClock)

	if config != nil {
		client = TestClient(config[0], fakeClock, NewRequester(config[0]))
	} else {
		singleConfig := Config{AppName: "CF CLI API V3 Test", AppVersion: "Unknown"}
		client = TestClient(
			singleConfig,
			fakeClock,
			NewRequester(singleConfig),
		)
	}
	client.TargetCF(TargetSettings{
		SkipSSLValidation: true,
		URL:               server.URL(),
	})

	return client, fakeClock
}
