package integrationtest_test

import (
	"testing"
)

func TestIntegrationtest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hydra Broker Test Suite")
}
