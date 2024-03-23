package composite_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWrappers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Composite Actors Suite")
}
