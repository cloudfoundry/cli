package command_factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCommandFactory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Command Factory Suite")
}
