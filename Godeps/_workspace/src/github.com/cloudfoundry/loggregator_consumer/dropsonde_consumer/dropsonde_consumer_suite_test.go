package dropsonde_consumer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDropsondeConsumer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DropsondeConsumer Suite")
}
