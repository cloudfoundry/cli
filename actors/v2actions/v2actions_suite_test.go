package v2actions_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestV2actions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V2 Actions Suite")
}
