package random_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRandom(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Random Suite")
}
