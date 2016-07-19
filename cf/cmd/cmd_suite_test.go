package cmd_test

import (
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cli/testhelpers/pluginbuilder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)

	SetDefaultEventuallyTimeout(2 * time.Second)

	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "fixtures", "plugins"), "test_1")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "fixtures", "plugins"), "test_2")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "fixtures", "plugins"), "test_with_push")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "fixtures", "plugins"), "test_with_push_short_name")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "fixtures", "plugins"), "test_with_help")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "fixtures", "plugins"), "my_say")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "fixtures", "plugins"), "call_core_cmd")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "fixtures", "plugins"), "input")

	//compile plugin examples to ensure they're up to date
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "plugin_examples"), "basic_plugin")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "plugin_examples"), "echo")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "plugin_examples"), "interactive")

	RunSpecs(t, "Cmd Suite")
}
