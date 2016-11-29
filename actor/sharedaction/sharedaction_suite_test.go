package sharedaction_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSharedAction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shared Actions Suite")
}
