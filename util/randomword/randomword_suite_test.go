package randomword_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRandomword(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Randomword Suite")
}
