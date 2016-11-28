package main_test

import (
	"code.cloudfoundry.org/cli/util/testhelpers/pluginbuilder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)

	pluginbuilder.BuildTestBinary(".", "call_cli_cmd")

	RunSpecs(t, "Main Suite")
}
