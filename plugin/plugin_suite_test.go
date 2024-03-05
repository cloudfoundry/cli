package plugin_test

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/cf/util/testhelpers/pluginbuilder"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	pluginbuilder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins"), "test_1")
	RunSpecs(t, "Plugin Suite")
}
