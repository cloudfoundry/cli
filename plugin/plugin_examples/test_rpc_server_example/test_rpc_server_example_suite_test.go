package main_test

import (
	"testing"
	"time"

	"code.cloudfoundry.org/cli/v9/cf/util/testhelpers/pluginbuilder"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTestRpcServerExample(t *testing.T) {
	RegisterFailHandler(Fail)

	pluginbuilder.BuildTestBinary("", "test_rpc_server_example")

	RunSpecs(t, "Test RPC Server Example Suite")
}

var _ = BeforeEach(func() {
	SetDefaultEventuallyTimeout(3 * time.Second)
})
