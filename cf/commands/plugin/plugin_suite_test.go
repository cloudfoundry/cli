package plugin_test

import (
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/commands/plugin"
	"github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/testhelpers/pluginbuilder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPlugin(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config)

	_ = plugin.Plugins{}

	RegisterFailHandler(Fail)

	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins"), "test_with_help")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins"), "test_with_orgs")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins"), "test_with_orgs_short_name")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins"), "test_with_push")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins"), "test_with_push_short_name")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins"), "test_1")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins"), "test_2")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins"), "empty_plugin")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins"), "alias_conflicts")

	RunSpecs(t, "Plugin Suite")
}
