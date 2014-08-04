package broker_builder_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBrokerBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BrokerBuilder Suite")
}
