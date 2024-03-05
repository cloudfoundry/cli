package selfcontained_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/v7/selfcontained/fake"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("cf api", func() {
	var apiConfig fake.CFAPIConfig

	BeforeEach(func() {
		apiConfig = fake.CFAPIConfig{
			Routes: map[string]fake.Response{
				"GET /": {Code: http.StatusOK, Body: ccv3.Info{}},
			},
		}
		apiServer.SetConfiguration(apiConfig)
	})

	JustBeforeEach(func() {
		Eventually(helpers.CF("api", apiServer.URL())).Should(gexec.Exit(0))
	})

	It("disables cf-on-k8s in config", func() {
		Expect(loadConfig().CFOnK8s.Enabled).To(BeFalse())
	})

	When("pointed to cf-on-k8s", func() {
		BeforeEach(func() {
			apiConfig.Routes["GET /"] = fake.Response{
				Code: http.StatusOK, Body: ccv3.Info{CFOnK8s: true},
			}
			apiServer.SetConfiguration(apiConfig)
		})

		It("enables cf-on-k8s in config", func() {
			Expect(loadConfig().CFOnK8s.Enabled).To(BeTrue())
		})
	})

	When("already logged into a cf-on-k8s", func() {
		BeforeEach(func() {
			helpers.SetConfig(func(config *configv3.Config) {
				config.ConfigFile.CFOnK8s.Enabled = true
				config.ConfigFile.CFOnK8s.AuthInfo = "something"
			})
		})

		It("disables cf-on-k8s in config and clears the auth-info", func() {
			Expect(loadConfig().CFOnK8s).To(Equal(configv3.CFOnK8s{
				Enabled:  false,
				AuthInfo: "",
			}))
		})

		When("pointed to cf-on-k8s", func() {
			BeforeEach(func() {
				apiConfig.Routes["GET /"] = fake.Response{
					Code: http.StatusOK, Body: ccv3.Info{CFOnK8s: true},
				}
				apiServer.SetConfiguration(apiConfig)
			})

			It("clears the auth-info", func() {
				Expect(loadConfig().CFOnK8s).To(Equal(configv3.CFOnK8s{
					Enabled:  true,
					AuthInfo: "",
				}))
			})
		})
	})
})

var _ = Describe("cf api --unset", func() {
	BeforeEach(func() {
		helpers.SetConfig(func(config *configv3.Config) {
			config.ConfigFile.CFOnK8s.Enabled = true
			config.ConfigFile.CFOnK8s.AuthInfo = "something"
		})
	})

	JustBeforeEach(func() {
		Eventually(helpers.CF("api", "--unset")).Should(gexec.Exit(0))
	})

	It("disables cf-on-k8s in config and clears the auth-info", func() {
		Expect(loadConfig().CFOnK8s).To(Equal(configv3.CFOnK8s{
			Enabled:  false,
			AuthInfo: "",
		}))
	})
})
