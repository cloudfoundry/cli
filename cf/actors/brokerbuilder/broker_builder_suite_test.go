package brokerbuilder_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBrokerBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BrokerBuilder Suite")
}
