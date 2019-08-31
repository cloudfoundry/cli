package pushmanifestparser_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestManifestparser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Manifest ParsedManifest Suite")
}
