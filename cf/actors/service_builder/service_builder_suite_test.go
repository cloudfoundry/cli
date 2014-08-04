package service_builder_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestServiceBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ServiceBuilder Suite")
}
