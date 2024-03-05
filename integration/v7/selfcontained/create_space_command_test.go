package selfcontained_test

import (
	"net/http"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/v7/selfcontained/fake"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

var _ = Describe("Create Space Command", func() {
	Describe("CF-on-k8s", func() {
		var (
			kubeConfigPath string
			kubeConfig     apiv1.Config
			session        *gexec.Session
			apiConfig      fake.CFAPIConfig
		)

		BeforeEach(func() {
			helpers.SetConfig(func(config *configv3.Config) {
				config.ConfigFile.Target = apiServer.URL()
				config.ConfigFile.CFOnK8s.Enabled = true
				config.ConfigFile.CFOnK8s.AuthInfo = "my-auth"
				config.ConfigFile.TargetedOrganization.GUID = "org-guid"
				config.ConfigFile.TargetedOrganization.Name = "my-org"
			})

			apiConfig = fake.CFAPIConfig{
				Routes: map[string]fake.Response{
					"POST /v3/spaces": {Code: http.StatusCreated},
					"POST /v3/roles":  {Code: http.StatusCreated},
					"GET /whoami": {
						Code: http.StatusOK, Body: map[string]interface{}{
							"name": "my-user",
							"kind": "User",
						},
					},
				},
			}
			apiServer.SetConfiguration(apiConfig)

			kubeConfig = apiv1.Config{
				Kind:       "Config",
				APIVersion: "v1",
				AuthInfos: []apiv1.NamedAuthInfo{
					{
						Name: "my-auth",
						AuthInfo: apiv1.AuthInfo{
							Token: "foo",
						},
					},
				},
				Clusters: []apiv1.NamedCluster{
					{
						Name: "my-cluster",
						Cluster: apiv1.Cluster{
							Server: "https://example.org",
						},
					},
				},
				Contexts: []apiv1.NamedContext{
					{
						Name: "my-context",
						Context: apiv1.Context{
							Cluster:   "my-cluster",
							AuthInfo:  "my-auth",
							Namespace: "my-namespace",
						},
					},
				},
				CurrentContext: "my-context",
			}

			kubeConfigPath := filepath.Join(homeDir, ".kube", "config")
			storeKubeConfig(kubeConfig, kubeConfigPath)

			env = helpers.CFEnv{
				EnvVars: map[string]string{
					"KUBECONFIG": kubeConfigPath,
				},
			}
		})

		JustBeforeEach(func() {
			session = helpers.CustomCF(env, "create-space", "my-space")
		})

		AfterEach(func() {
			Expect(os.RemoveAll(kubeConfigPath)).To(Succeed())
		})

		It("creates a role using the name from the /whoami endpoint", func() {
			Eventually(session).Should(gexec.Exit(0))

			Expect(session).To(gbytes.Say("Assigning role SpaceManager to user my-user in org my-org"))
			Expect(session).To(gbytes.Say("Assigning role SpaceDeveloper to user my-user in org my-org"))
		})
	})
})
