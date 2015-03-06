package noaa_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestNoaa(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Noaa Suite")
}
