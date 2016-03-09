package cloud_controller_gateway

import (
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/trace/fakes"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
)

func NewTestCloudControllerGateway(configRepo core_config.Reader) net.Gateway {
	fakeLogger := new(fakes.FakePrinter)
	return net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{}, fakeLogger)
}
