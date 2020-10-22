package batcher_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Batcher Suite")
}
