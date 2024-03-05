package configuration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConfiguration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Configuration Suite")
}
