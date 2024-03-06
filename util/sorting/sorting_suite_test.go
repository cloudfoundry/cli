package sorting_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSorting(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sorting Suite")
}
