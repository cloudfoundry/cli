package shared_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestShared(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V6 Command's Shared Suite")
}
