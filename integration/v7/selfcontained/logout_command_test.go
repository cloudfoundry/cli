package selfcontained_test

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("cf logout", func() {
	BeforeEach(func() {
		helpers.SetConfig(func(config *configv3.Config) {
			config.ConfigFile.CFOnK8s.Enabled = true
			config.ConfigFile.CFOnK8s.AuthInfo = "something"
		})
	})

	JustBeforeEach(func() {
		Eventually(helpers.CF("logout")).Should(gexec.Exit(0))
	})

	It("clears the auth-info", func() {
		Expect(loadConfig().CFOnK8s).To(Equal(configv3.CFOnK8s{
			Enabled:  true,
			AuthInfo: "",
		}))
	})
})
