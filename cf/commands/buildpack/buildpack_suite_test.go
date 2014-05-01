package buildpack_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBuildpack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Buildpack Suite")
}
