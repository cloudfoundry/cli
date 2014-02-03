package servicebroker_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestServicebroker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Servicebroker Suite")
}
