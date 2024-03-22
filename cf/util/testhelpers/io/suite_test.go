package io_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestIOHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IO Helpers Suite")
}
