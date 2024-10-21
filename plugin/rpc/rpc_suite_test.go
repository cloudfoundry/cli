package rpc_test

import (
	"code.cloudfoundry.org/cli/v7/plugin/rpc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

var rpcService *rpc.CliRpcService

func TestRpc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RPC Suite")
}
