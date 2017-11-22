package ykk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestYkk(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "YKK Suite")
}
