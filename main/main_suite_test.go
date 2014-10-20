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
	plugin_builder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins"), "test_with_help")
	plugin_builder.BuildTestBinary(filepath.Join("..", "fixtures", "plugins"), "my_say")
	RunSpecs(t, "Main Suite")
}
