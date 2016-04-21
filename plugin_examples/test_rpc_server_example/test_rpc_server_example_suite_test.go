package main_test

import (
	"github.com/cloudfoundry/cli/testhelpers/pluginbuilder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTestRPCServerExample(t *testing.T) {
	RegisterFailHandler(Fail)

	pluginbuilder.BuildTestBinary("", "test_rpc_server_example")

	RunSpecs(t, "TestRPCServerExample Suite")
}
