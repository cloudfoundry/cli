package command_factory_test

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	testPluginConfig "github.com/cloudfoundry/cli/cf/configuration/plugin_config/fakes"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/plugin/rpc"
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
		pluginConfig := &testPluginConfig.FakePluginConfiguration{}
		repoLocator := api.NewRepositoryLocator(config, map[string]net.Gateway{
			"auth":             net.NewUAAGateway(config, fakeUI),
			"cloud-controller": net.NewCloudControllerGateway(config, time.Now, fakeUI),
			"uaa":              net.NewUAAGateway(config, fakeUI),
		})

		rpcService, _ := rpc.NewRpcService(nil, nil, nil, nil)
		factory = NewFactory(fakeUI, config, manifestRepo, repoLocator, pluginConfig, rpcService)
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

			Expect(matchingCount).To(Equal(1), "this file has no corresponding command: "+info.Name())
			return nil
		})

		Expect(err).NotTo(HaveOccurred())
	})
	Describe("GetByCmdName", func() {
		It("returns the cmd if it exists", func() {
			cmd, err := factory.GetByCmdName("push")
			Expect(cmd).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("can find commands by short name", func() {
			cmd, err := factory.GetByCmdName("p")
			Expect(cmd).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error if it does not exist", func() {
			cmd, err := factory.GetByCmdName("FOOBARRRRR")
			Expect(cmd).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("CheckIfCoreCmdExists", func() {
		It("returns true if the cmd exists", func() {
			exists := factory.CheckIfCoreCmdExists("push")
			Expect(exists).To(BeTrue())
		})

		It("retruns true if the cmd short name exists", func() {
			exists := factory.CheckIfCoreCmdExists("p")
			Expect(exists).To(BeTrue())
		})

		It("returns an error if it does not exist", func() {
			exists := factory.CheckIfCoreCmdExists("FOOOOBARRRR")
			Expect(exists).To(BeFalse())
		})
	})

	Describe("GetCommandFlags", func() {
		It("returns a list of flags for the command", func() {
			flags := factory.GetCommandFlags("push")
			Expect(contains(flags, "b")).To(Equal(true))
			Expect(contains(flags, "c")).To(Equal(true))
			Expect(contains(flags, "no-hostname")).To(Equal(true))
		})
	})

	Describe("GetCommandTotalArgs", func() {
		It("returns the total number of argument required by the command ", func() {
			totalArgs, err := factory.GetCommandTotalArgs("create-buildpack")
			Expect(err).ToNot(HaveOccurred())
			Expect(totalArgs).To(Equal(3))
		})

		It("returns an error if command does not exist", func() {
			_, err := factory.GetCommandTotalArgs("not-a-command")
			Expect(err).To(HaveOccurred())
		})
	})
})

func contains(ary []string, item string) bool {
	for _, v := range ary {
		if v == item {
			return true
		}
	}
	return false
}
