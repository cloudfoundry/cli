package commandregistry_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCommandRegistry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CommandRegistry Suite")
}
