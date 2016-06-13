package actors_test

import (
	"github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestActors(t *testing.T) {
	i18n.T = i18n.Init(configuration.NewRepositoryWithDefaults())
	RegisterFailHandler(Fail)
	RunSpecs(t, "Actors Suite")
}
