package fixtures_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFixtures(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fixtures Suite")
}
