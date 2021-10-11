package selfcontained_test

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/v7/selfcontained/fake"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	homeDir   string
	apiServer *fake.CFAPI
)

func TestSelfcontained(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Selfcontained Suite")
}

var _ = BeforeEach(func() {
	homeDir = helpers.SetHomeDir()
	apiServer = fake.NewCFAPI()
	helpers.SetConfig(func(config *configv3.Config) {
		config.ConfigFile.Target = apiServer.URL()
	})
})

var _ = AfterEach(func() {
	apiServer.Close()
	helpers.DestroyHomeDir(homeDir)
})

func loadConfig() configv3.JSONConfig {
	rawConfig, err := ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
	Expect(err).NotTo(HaveOccurred())

	var configFile configv3.JSONConfig
	Expect(json.Unmarshal(rawConfig, &configFile)).To(Succeed())

	return configFile
}
