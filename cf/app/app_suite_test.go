package app_test

import (
	"github.com/cloudfoundry/cli/cf/commands/buildpack"
	"github.com/cloudfoundry/cli/cf/commands/domain"
	"github.com/cloudfoundry/cli/cf/commands/organization"
	"github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/i18n/detection"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/testhelpers/plugin_builder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"path/filepath"

	"testing"
)

func TestApp(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config, &detection.JibberJabberDetector{})

	//make a reference to something in cf/commands/domain, so all init() in the directory will run
	_ = domain.CreateDomain{}
	_ = buildpack.ListBuildpacks{}
	_ = organization.ListOrgs{}

	RegisterFailHandler(Fail)
	plugin_builder.BuildTestBinary(filepath.Join("..", "..", "fixtures", "plugins"), "test_1")
	plugin_builder.BuildTestBinary(filepath.Join("..", "..", "fixtures", "plugins"), "test_2")
	RunSpecs(t, "App Suite")
}
