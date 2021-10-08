package selfcontained_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/v7/selfcontained/fake"
	"code.cloudfoundry.org/cli/util/configv3"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	homeDir   string
	apiServer *fake.CFAPI
	env       helpers.CFEnv
	token     []byte
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

	keyPair, err := rsa.GenerateKey(rand.Reader, 2048)
	Expect(err).NotTo(HaveOccurred())

	jwt := jws.NewJWT(jws.Claims{
		"exp": time.Now().Add(time.Hour).Unix(),
	}, crypto.SigningMethodRS256)
	token, err = jwt.Serialize(keyPair)
	Expect(err).NotTo(HaveOccurred())
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
