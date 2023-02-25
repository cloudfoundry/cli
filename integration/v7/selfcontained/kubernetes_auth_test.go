package selfcontained_test

import (
	"net/http"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/v7/selfcontained/fake"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
)

var _ = Describe("auth-provider", func() {
	var (
		apiConfig  fake.CFAPIConfig
		kubeConfig apiv1.Config
	)

	BeforeEach(func() {
		apiConfig = fake.CFAPIConfig{
			Routes: map[string]fake.Response{
				"GET /v3/apps": {
					Code: http.StatusOK, Body: map[string]interface{}{
						"pagination": map[string]interface{}{},
						"resources":  []resources.Application{},
					},
				},
				"GET /whoami": {
					Code: http.StatusOK, Body: map[string]interface{}{
						"name": "my-user",
						"kind": "User",
					},
				},
			},
		}
		apiServer.SetConfiguration(apiConfig)
		helpers.SetConfig(func(config *configv3.Config) {
			config.ConfigFile.Target = apiServer.URL()
			config.ConfigFile.CFOnK8s.Enabled = true
			config.ConfigFile.CFOnK8s.AuthInfo = "my-user"
			config.ConfigFile.TargetedOrganization = configv3.Organization{
				GUID: "my-org",
				Name: "My Org",
			}

			config.ConfigFile.TargetedSpace = configv3.Space{
				GUID: "my-space",
				Name: "My Space",
			}
		})

		kubeConfig = apiv1.Config{
			Kind:       "Config",
			APIVersion: "v1",
			AuthInfos: []apiv1.NamedAuthInfo{
				{
					Name: "my-user",
					AuthInfo: apiv1.AuthInfo{
						AuthProvider: &apiv1.AuthProviderConfig{
							Name: "oidc",
							Config: map[string]string{
								"id-token":       string(token),
								"idp-issuer-url": "-",
								"client-id":      "-",
							},
						},
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
						AuthInfo:  "my-auth-info",
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
		Eventually(helpers.CustomCF(env, "apps")).Should(gexec.Exit(0))
	})

	It("sends the Bearer token in the Authorization header", func() {
		reqs := apiServer.ReceivedRequests()["GET /v3/apps"]
		Expect(reqs).To(HaveLen(1))
		Expect(reqs[0].Header).To(HaveKeyWithValue("Authorization", ConsistOf("Bearer "+string(token))))
	})
})
