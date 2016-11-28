package v2action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestV2Action(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V2 Actions Suite")
}
