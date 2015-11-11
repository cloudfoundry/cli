package main_test

import (
	"github.com/cloudfoundry/cli/testhelpers/plugin_builder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTestRpcServerExample(t *testing.T) {
	RegisterFailHandler(Fail)

	plugin_builder.BuildTestBinary("", "test_rpc_server_example")

	RunSpecs(t, "TestRpcServerExample Suite")
}
