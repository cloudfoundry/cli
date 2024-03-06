package matchers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMatchers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Matchers Suite")
}
