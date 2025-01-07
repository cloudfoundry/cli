package plugin_test

import (
	"path/filepath"
	"testing"

	"code.cloudfoundry.org/cli/v8/cf/util/testhelpers/pluginbuilder"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	pluginbuilder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins", "test_1"), "test_1")
	RunSpecs(t, "Plugin Suite")
}
