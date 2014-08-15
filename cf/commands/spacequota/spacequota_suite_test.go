package spacequota_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSpacequota(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spacequota Suite")
}
