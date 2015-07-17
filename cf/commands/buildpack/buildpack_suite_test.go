package buildpack_test

import (
	"github.com/cloudfoundry/cli/cf/commands/buildpack"
	"github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/i18n/detection"
	"github.com/cloudfoundry/cli/testhelpers/configuration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBuildpack(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config, &detection.JibberJabberDetector{})

	//make a reference to something in cf/commands/domain, so all init() in the directory will run
	_ = buildpack.ListBuildpacks{}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Buildpack Suite")
}
