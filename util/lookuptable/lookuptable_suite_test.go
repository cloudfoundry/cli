package lookuptable_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLookuptable(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lookuptable Suite")
}
