package commands_test

import (
	"code.cloudfoundry.org/cli/cf/commands"
	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/util/testhelpers/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"testing"
)

func TestCommands(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config)

	_ = commands.API{}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Commands Suite")
}

var _ = BeforeEach(func() {
	log.SetLevel(log.PanicLevel)
})

type passingRequirement struct {
	Name string
}

func (r passingRequirement) Execute() error {
	return nil
}
