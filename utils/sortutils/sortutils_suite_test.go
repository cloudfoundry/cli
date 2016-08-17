package sortutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSortutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sortutils Suite")
}
