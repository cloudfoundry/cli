package unique_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUnique(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unique Suite")
}
