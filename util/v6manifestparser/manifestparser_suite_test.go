package v6manifestparser_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestManifestparser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Manifest Parser Suite")
}
