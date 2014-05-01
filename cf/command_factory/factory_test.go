package command_factory_test

import (
	. "github.com/cloudfoundry/cli/cf/command_factory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
)

var _ = Describe("factory", func() {
	var (
		factory Factory
	)

	BeforeEach(func() {
		fakeUI := &testterm.FakeUI{}
		config := testconfig.NewRepository()
		manifestRepo := manifest.NewManifestDiskRepository()
		repoLocator := api.NewRepositoryLocator(config, map[string]net.Gateway{
			"auth":             net.NewUAAGateway(config),
			"cloud-controller": net.NewCloudControllerGateway(config),
			"uaa":              net.NewUAAGateway(config),
		})

		factory = NewFactory(fakeUI, config, manifestRepo, repoLocator)
	})

	It("provides the metadata for its commands", func() {
		commands := factory.CommandMetadatas()

		err := filepath.Walk("../commands", func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, "_test.go") || info.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".test") {
				return nil
			}

			extension := filepath.Ext(info.Name())
			expectedCommandName := strings.Replace(info.Name()[0:len(info.Name())-len(extension)], "_", "-", -1)

			matchingCount := 0
			for _, command := range commands {
				if command.Name == expectedCommandName {
					matchingCount++
				}
			}

			Expect(matchingCount).To(Equal(1), "this command is not tested: "+info.Name())
			return nil
		})

		Expect(err).NotTo(HaveOccurred())
	})
})
