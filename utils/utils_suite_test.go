package utils_test

import (
	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/testhelpers/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUtils(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}
