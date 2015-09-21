package multiopt_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMultiopt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multiopt Suite")
}
