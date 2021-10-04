package selfcontained_test

import (
	"testing"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var homeDir string

func TestSelfcontained(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Selfcontained Suite")
}

var _ = BeforeEach(func() {
	homeDir = helpers.SetHomeDir()
})

var _ = AfterEach(func() {
	helpers.DestroyHomeDir(homeDir)
})
