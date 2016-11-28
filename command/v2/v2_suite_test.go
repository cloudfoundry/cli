package v2_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestV2(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V2 Command Suite")
}
