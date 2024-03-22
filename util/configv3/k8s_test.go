package configv3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/util/configv3"
)

var _ = Describe("K8s", func() {
	var config configv3.Config

	BeforeEach(func() {
		config = configv3.Config{}
	})

	Describe("IsCFOnK8s", func() {
		It("returns false by default", func() {
			Expect(config.IsCFOnK8s()).To(BeFalse())
		})

		When("the config is pointed to cf-on-k8s", func() {
			BeforeEach(func() {
				config.ConfigFile.CFOnK8s = configv3.CFOnK8s{
					Enabled: true,
				}
			})

			It("returns true", func() {
				Expect(config.IsCFOnK8s()).To(BeTrue())
			})
		})
	})
})
