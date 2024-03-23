package integrationtest_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIntegrationtest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hydra Broker Test Suite")
}
