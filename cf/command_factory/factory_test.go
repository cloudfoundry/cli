package command_factory_test

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/command_factory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			"cloud-controller": net.NewCloudControllerGateway(config, time.Now),
			"uaa":              net.NewUAAGateway(config),
		})

		factory = NewFactory(fakeUI, config, manifestRepo, repoLocator)
	})

	It("provides the metadata for its commands", func() {
		commands := factory.CommandMetadatas()

		suffixesToIgnore := []string{
			"i18n_init.go", // ignore all i18n initializers
			"_test.go",     // ignore test files
			".test",        // ignore generated .test (temporary files)
			"#",            // emacs autosave files
		}

		err := filepath.Walk("../commands", func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			for _, suffix := range suffixesToIgnore {
				if strings.HasSuffix(path, suffix) {
					return nil
				}
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
