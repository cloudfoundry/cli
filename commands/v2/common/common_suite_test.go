package common_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestInternal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V2 Commands Common Suite")
}
