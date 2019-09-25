// +build V7

package rpc_test

import (
	"code.cloudfoundry.org/cli/plugin/v7/rpc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var rpcService *rpc.CliRpcService

func TestRpc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RPC Suite")
}
