package generic_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGeneric(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generic Suite")
}
