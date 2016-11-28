package io_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestIOHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IO Helpers Suite")
}
