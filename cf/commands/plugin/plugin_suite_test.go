package plugin_test

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/v8/cf/commands/plugin"
	"code.cloudfoundry.org/cli/v8/cf/i18n"
	"code.cloudfoundry.org/cli/v8/cf/util/testhelpers/configuration"
	"code.cloudfoundry.org/cli/v8/cf/util/testhelpers/pluginbuilder"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPlugin(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config)

	_ = plugin.Plugins{}

	RegisterFailHandler(Fail)

	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins", "test_with_help"), "test_with_help")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins", "test_with_orgs"), "test_with_orgs")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins", "test_with_orgs_short_name"), "test_with_orgs_short_name")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins", "test_with_push"), "test_with_push")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins", "test_with_push_short_name"), "test_with_push_short_name")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins", "test_1"), "test_1")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins", "test_2"), "test_2")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins", "empty_plugin"), "empty_plugin")
	pluginbuilder.BuildTestBinary(filepath.Join("..", "..", "..", "fixtures", "plugins", "alias_conflicts"), "alias_conflicts")

	RunSpecs(t, "Plugin Suite")
}
