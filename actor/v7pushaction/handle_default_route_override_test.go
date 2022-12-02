package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleDefaultRouteOverride", func() {
	var (
		originalManifest    manifestparser.Manifest
		transformedManifest manifestparser.Manifest
		overrides           FlagOverrides
		executeErr          error
	)

	BeforeEach(func() {
		originalManifest = manifestparser.Manifest{}
		overrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = HandleDefaultRouteOverride(originalManifest, overrides)
	})

	When("the manifest has the no-route field", func() {
		BeforeEach(func() {
			originalManifest = manifestparser.Manifest{
				Applications: []manifestparser.Application{{NoRoute: true}},
			}
		})
		It("does not add default route", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(transformedManifest).To(Equal(manifestparser.Manifest{
				Applications: []manifestparser.Application{{NoRoute: true}},
			}))
		})

	})

	When("the manifest has the random-route field", func() {
		BeforeEach(func() {
			originalManifest = manifestparser.Manifest{
				Applications: []manifestparser.Application{{RandomRoute: true}},
			}
		})
		It("does not add default route", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(transformedManifest).To(Equal(manifestparser.Manifest{
				Applications: []manifestparser.Application{{RandomRoute: true}},
			}))

		})

	})

	// CLI doesnt know about the routes field but CAPI ignores default route if routes is specified
	// so we are ok adding default route even with the presence of a routes field

	When("the manifest has no routing fields", func() {
		BeforeEach(func() {
			originalManifest = manifestparser.Manifest{
				Applications: []manifestparser.Application{{}},
			}
		})
		It("does add default route", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(transformedManifest).To(Equal(manifestparser.Manifest{
				Applications: []manifestparser.Application{{DefaultRoute: true}},
			}))

		})

	})

})
