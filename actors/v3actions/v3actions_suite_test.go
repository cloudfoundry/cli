package v3actions_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestV3actions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V3actions Suite")
}
