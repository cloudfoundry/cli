package v7_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestV7(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V7 Suite")
}
