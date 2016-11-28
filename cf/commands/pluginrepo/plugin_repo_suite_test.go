package pluginrepo_test

import (
	"code.cloudfoundry.org/cli/cf/commands/pluginrepo"
	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/util/testhelpers/configuration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPluginRepo(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config)

	_ = pluginrepo.RepoPlugins{}

	RegisterFailHandler(Fail)
	RunSpecs(t, "PluginRepo Suite")
}
