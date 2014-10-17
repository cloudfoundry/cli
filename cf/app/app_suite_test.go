package app_test

import (
	"github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/testhelpers/plugin_builder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"path/filepath"

	"testing"
)

func TestApp(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config)

	RegisterFailHandler(Fail)
	plugin_builder.BuildTestBinary(filepath.Join("..", "..", "fixtures", "plugins"), "test_1")
	plugin_builder.BuildTestBinary(filepath.Join("..", "..", "fixtures", "plugins"), "test_2")
	RunSpecs(t, "App Suite")
}
