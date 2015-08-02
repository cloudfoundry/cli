package help_test

import (
	"github.com/cloudfoundry/cli/commands_loader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHelp(t *testing.T) {
	RegisterFailHandler(Fail)

	commands_loader.Load()

	RunSpecs(t, "Help Suite")
}
