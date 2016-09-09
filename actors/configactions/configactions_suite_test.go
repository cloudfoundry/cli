package configactions_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConfigactions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Configactions Suite")
}
