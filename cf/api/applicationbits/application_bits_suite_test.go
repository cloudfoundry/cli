package applicationbits_test

import (
	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/util/testhelpers/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestApplicationBits(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config)

	RegisterFailHandler(Fail)
	RunSpecs(t, "ApplicationBits Suite")
}
