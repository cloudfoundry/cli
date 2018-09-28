package shared_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestShared(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V2 Command's Shared Suite")
}
