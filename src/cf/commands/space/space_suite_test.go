package space_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSpace(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Space Suite")
}
