package help_test

import (
	"github.com/cloudfoundry/cli/cf/commandsloader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHelp(t *testing.T) {
	RegisterFailHandler(Fail)

	commandsloader.Load()

	RunSpecs(t, "Help Suite")
}
