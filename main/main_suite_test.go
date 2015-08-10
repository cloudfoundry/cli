package main_test

import (
	"path/filepath"

	"github.com/cloudfoundry/cli/testhelpers/plugin_builder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)

	plugin_builder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins"), "test_1")
	plugin_builder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins"), "test_2")
	plugin_builder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins"), "test_with_push")
	plugin_builder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins"), "test_with_push_short_name")
	plugin_builder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins"), "test_with_help")
	plugin_builder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins"), "my_say")
	plugin_builder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins"), "call_core_cmd")
	plugin_builder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins"), "input")
	plugin_builder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins"), "panics")

	//compile plugin examples to ensure they're up to date
	plugin_builder.BuildTestBinary(filepath.Join("..", "plugin_examples"), "basic_plugin")
	plugin_builder.BuildTestBinary(filepath.Join("..", "plugin_examples"), "echo")
	plugin_builder.BuildTestBinary(filepath.Join("..", "plugin_examples"), "interactive")

	RunSpecs(t, "Main Suite")
}
