package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/v8/actor/v7pushaction"
	"code.cloudfoundry.org/cli/v8/util/manifestparser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleCNBCredentialsOverride", func() {
	var (
		originalManifest    manifestparser.Manifest
		transformedManifest manifestparser.Manifest
		overrides           FlagOverrides
		executeErr          error
	)

	BeforeEach(func() {
		originalManifest = manifestparser.Manifest{
			Applications: []manifestparser.Application{{}},
		}
		overrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = HandleCNBCredentialsOverride(originalManifest, overrides)
	})

	When("the cnb credentials are present", func() {
		BeforeEach(func() {
			overrides.CNBCredentials = map[string]interface{}{
				"foo": "bar",
			}
		})

		It("add it to the raw manifest", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(transformedManifest).To(Equal(manifestparser.Manifest{
				Applications: []manifestparser.Application{{
					RemainingManifestFields: map[string]interface{}{
						"cnb-credentials": map[string]interface{}{
							"foo": "bar",
						},
					},
				}},
			}))
		})

	})

	When("the credentials are not present", func() {
		BeforeEach(func() {
			overrides.CNBCredentials = nil
		})

		It("does not add it to the raw manifest", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(transformedManifest).To(Equal(manifestparser.Manifest{
				Applications: []manifestparser.Application{{}},
			}))

		})

	})
})
