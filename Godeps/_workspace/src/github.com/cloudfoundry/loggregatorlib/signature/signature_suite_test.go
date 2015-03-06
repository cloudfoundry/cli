package signature

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSignature(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Signature Suite")
}
