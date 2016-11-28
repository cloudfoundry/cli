package actors_test

import (
	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/util/testhelpers/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestActors(t *testing.T) {
	i18n.T = i18n.Init(configuration.NewRepositoryWithDefaults())
	RegisterFailHandler(Fail)
	RunSpecs(t, "Actors Suite")
}
