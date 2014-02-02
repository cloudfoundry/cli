package requirements_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRequirements(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Requirements Suite")
}
