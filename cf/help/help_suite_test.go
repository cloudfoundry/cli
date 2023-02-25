package help_test

import (
	"code.cloudfoundry.org/cli/cf/commandsloader"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHelp(t *testing.T) {
	RegisterFailHandler(Fail)

	commandsloader.Load()

	RunSpecs(t, "Help Suite")
}
