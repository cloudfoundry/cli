package rpc_test

import (
	"github.com/cloudfoundry/cli/plugin/rpc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var rpcService *rpc.CliRPCService

func TestRPC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RPC Suite")
}
