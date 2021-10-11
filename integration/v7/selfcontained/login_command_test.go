package selfcontained_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/v7/selfcontained/fake"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"gopkg.in/yaml.v2"
	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

var _ = Describe("LoginCommand", func() {
	Describe("CF-on-k8s", func() {
		var (
			stdin          *gbytes.Buffer
			kubeConfigPath string
			kubeConfig     apiv1.Config
			env            helpers.CFEnv
			session        *gexec.Session
			loginArgs      []string
			apiConfig      fake.CFAPIConfig
		)

		BeforeEach(func() {
			loginArgs = []string{"login"}
			helpers.SetConfig(func(config *configv3.Config) {
				config.ConfigFile.Target = apiServer.URL()
				config.ConfigFile.CFOnK8s.Enabled = true
			})

			apiConfig = fake.CFAPIConfig{
				Routes: map[string]fake.Response{
					"GET /": {Code: http.StatusOK, Body: ccv3.Info{CFOnK8s: true}},
					"GET /v3/organizations": {
						Code: http.StatusOK, Body: map[string]interface{}{
							"pagination": map[string]interface{}{},
							"resources":  []resources.Organization{},
						},
					},
				},
			}
			apiServer.SetConfiguration(apiConfig)

			kubeConfig = apiv1.Config{
				Kind:        "Config",
				APIVersion:  "v1",
				Preferences: apiv1.Preferences{},
				Clusters: []apiv1.NamedCluster{
					{
						Name:    "cluster1",
						Cluster: apiv1.Cluster{},
					},
				},
				AuthInfos: []apiv1.NamedAuthInfo{
					{Name: "one", AuthInfo: apiv1.AuthInfo{Token: "asdf"}},
					{Name: "two", AuthInfo: apiv1.AuthInfo{Token: "fdsa"}},
				},
				Contexts: []apiv1.NamedContext{
					{
						Name: "ctx1",
						Context: apiv1.Context{
							Cluster:   "cluster1",
							AuthInfo:  "one",
							Namespace: "default",
						},
					},
				},
				CurrentContext: "ctx1",
			}
			kubeConfigPath := filepath.Join(homeDir, ".kube", "config")
			storeKubeConfig(kubeConfig, kubeConfigPath)

			stdin = gbytes.NewBuffer()
			_, wErr := fmt.Fprintf(stdin, "%d\n", 2)
			Expect(wErr).ToNot(HaveOccurred())

			env = helpers.CFEnv{
				Stdin: stdin,
				EnvVars: map[string]string{
					"KUBECONFIG": kubeConfigPath,
				},
			}
		})

		JustBeforeEach(func() {
			session = helpers.CustomCF(env, loginArgs...)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(kubeConfigPath)).To(Succeed())
		})

		It("prompts the user to select a user from the kube config file", func() {
			Eventually(session).Should(gbytes.Say("1. one"))
			Eventually(session).Should(gbytes.Say("2. two"))
			Eventually(session).Should(gbytes.Say("Choose your Kubernetes authentication info"))
			Eventually(session).Should(gbytes.Say("OK"))
			Eventually(session).Should(gexec.Exit(0))
		})

		It("sets the user into the configuration", func() {
			Eventually(session).Should(gexec.Exit(0))
			Expect(loadConfig().CFOnK8s.AuthInfo).To(Equal("two"))
		})

		It("displays the logged in user", func() {
			Eventually(session).Should(gbytes.Say("user:(\\s*)two"))
		})

		When("the kubeconfig contains no user information", func() {
			BeforeEach(func() {
				kubeConfig.AuthInfos = []apiv1.NamedAuthInfo{}
				storeKubeConfig(kubeConfig, filepath.Join(homeDir, ".kube", "config"))
			})

			It("displays an error", func() {
				Eventually(session.Err).Should(gbytes.Say("Unable to authenticate."))
				Eventually(session).Should(gbytes.Say("FAILED"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		When("providing -a flag without having targeted the api before", func() {
			BeforeEach(func() {
				helpers.SetConfig(func(config *configv3.Config) {
					config.ConfigFile.Target = ""
					config.ConfigFile.CFOnK8s = configv3.CFOnK8s{}
				})

				loginArgs = append(loginArgs, "-a", apiServer.URL())
			})

			It("displays the logged in user", func() {
				Eventually(session).Should(gbytes.Say("user:(\\s*)two"))
			})
		})
	})
})

func storeKubeConfig(kubeconfig apiv1.Config, kubeConfigPath string) {
	Expect(os.MkdirAll(filepath.Dir(kubeConfigPath), 0755)).To(Succeed())
	kubeConfigFile, err := os.OpenFile(kubeConfigPath, os.O_CREATE|os.O_WRONLY, 0755)
	Expect(kubeConfigFile.Truncate(0)).To(Succeed())
	Expect(err).NotTo(HaveOccurred())

	// we need to serialise the config to JSON as the Config type only has json annotations (and no yaml ones)
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
