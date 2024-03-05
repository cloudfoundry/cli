package selfcontained_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/v7/selfcontained/fake"
	"code.cloudfoundry.org/cli/util/configv3"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
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

func storeKubeConfig(kubeconfig apiv1.Config, kubeConfigPath string) {
	Expect(os.MkdirAll(filepath.Dir(kubeConfigPath), 0o755)).To(Succeed())
	kubeConfigFile, err := os.OpenFile(kubeConfigPath, os.O_CREATE|os.O_WRONLY, 0o755)
	Expect(kubeConfigFile.Truncate(0)).To(Succeed())
	Expect(err).NotTo(HaveOccurred())

	// we need to serialise the config to JSON as the Config type only has json annotations (and no yaml ones)
	// However, during json serialisation, byte arrays are base64 encoded which is not a desired side effect.
	// In order to address this, we base64 decode them in advance
	kubeconfig = base64DecodeClientCertByteArrays(kubeconfig)
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(kubeconfig)
	Expect(err).NotTo(HaveOccurred())

	var configmap map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &configmap)
	Expect(err).NotTo(HaveOccurred())

	// now we can save the config as yaml
	err = yaml.NewEncoder(kubeConfigFile).Encode(configmap)
	Expect(err).NotTo(HaveOccurred())
	Expect(kubeConfigFile.Close()).To(Succeed())
}

func base64DecodeClientCertByteArrays(kubeconfig apiv1.Config) apiv1.Config {
	decodedAuthInfos := []apiv1.NamedAuthInfo{}
	for _, authInfo := range kubeconfig.AuthInfos {
		if len(authInfo.AuthInfo.ClientCertificateData) > 0 {
			decodedCertData, err := base64.StdEncoding.DecodeString(string(authInfo.AuthInfo.ClientCertificateData))
			Expect(err).NotTo(HaveOccurred())
			authInfo.AuthInfo.ClientCertificateData = decodedCertData
		}
		if len(authInfo.AuthInfo.ClientKeyData) > 0 {
			decodedKeyData, err := base64.StdEncoding.DecodeString(string(authInfo.AuthInfo.ClientKeyData))
			Expect(err).NotTo(HaveOccurred())
			authInfo.AuthInfo.ClientKeyData = decodedKeyData
		}

		decodedAuthInfos = append(decodedAuthInfos, authInfo)
	}

	kubeconfig.AuthInfos = decodedAuthInfos
	return kubeconfig
}
