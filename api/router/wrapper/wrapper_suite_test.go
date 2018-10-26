package wrapper

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWrapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Router Wrapper Suite")
}
