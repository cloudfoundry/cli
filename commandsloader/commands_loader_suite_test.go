package commandsloader_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCommandsLoader(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "CommandsLoader Suite")
}
