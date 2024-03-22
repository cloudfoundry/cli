package configv3_test

import (
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("KubernetesUserConfig", func() {
	var (
		config configv3.KubernetesUserConfig
		err    error
	)

	BeforeEach(func() {
		config = configv3.KubernetesUserConfig{
			ConfigFile: &configv3.JSONConfig{
				CFOnK8s: configv3.CFOnK8s{
					AuthInfo: "kubernetes-user",
				},
			},
		}
	})

	Describe("CurrentUser", func() {
		var user configv3.User

		JustBeforeEach(func() {
			user, err = config.CurrentUser()
		})

		It("returns the configured auth-info", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(user).To(Equal(configv3.User{Name: "kubernetes-user"}))
		})
	})

	Describe("CurrentUserName", func() {
		var userName string

		JustBeforeEach(func() {
			userName, err = config.CurrentUserName()
		})

		It("returns the configured auth-info", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(userName).To(Equal("kubernetes-user"))
		})
	})
})
