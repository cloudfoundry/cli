package spaces_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSpaces(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spaces Suite")
}
