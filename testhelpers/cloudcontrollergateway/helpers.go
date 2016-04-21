package cloudcontrollergateway

import (
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/trace/tracefakes"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
)

func NewTestCloudControllerGateway(configRepo coreconfig.Reader) net.Gateway {
	fakeLogger := new(tracefakes.FakePrinter)
	return net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{}, fakeLogger)
}
