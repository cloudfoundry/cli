package application_test

import (
	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/util/testhelpers/configuration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestApplication(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Application Suite")
}

type passingRequirement struct {
	Name string
}

func (r passingRequirement) Execute() error {
	return nil
}
