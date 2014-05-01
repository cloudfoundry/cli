package loggregator_consumer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLoggregator_consumer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loggregator_consumer Suite")
}
