package buildpack_test

import (
	"code.cloudfoundry.org/cli/v7/cf/i18n"
	"code.cloudfoundry.org/cli/v7/cf/util/testhelpers/configuration"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBuildpack(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Buildpack Suite")
}
