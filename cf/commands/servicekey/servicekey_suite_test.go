package servicekey_test

import (
	"testing"

	"github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/testhelpers/configuration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestServicekey(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Servicekey Suite")
}
