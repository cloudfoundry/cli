package integrationtest_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegrationtest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hydra Broker Test Suite")
}
