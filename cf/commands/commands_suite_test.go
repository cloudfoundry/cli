package commands_test

import (
	"io"
	"log/slog"

	"code.cloudfoundry.org/cli/v9/cf/commands"
	"code.cloudfoundry.org/cli/v9/cf/i18n"
	"code.cloudfoundry.org/cli/v9/cf/util/testhelpers/configuration"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCommands(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config)

	_ = commands.API{}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Commands Suite")
}

var _ = BeforeSuite(func() {
	// Suppress log output during tests. This equates with setting to panic-level severity in older logging libraries
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
})

type passingRequirement struct {
	Name string
}

func (r passingRequirement) Execute() error {
	return nil
}
