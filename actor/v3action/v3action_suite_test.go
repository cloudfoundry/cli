package v3action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestV3Action(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V3 Actions Suite")
}
