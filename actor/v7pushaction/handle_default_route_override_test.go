package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/util/pushmanifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleDefualtRouteOverride", func() {
	var (
		originalManifest    pushmanifestparser.Manifest
		transformedManifest pushmanifestparser.Manifest
		overrides           FlagOverrides
		executeErr          error
	)

	BeforeEach(func() {
		originalManifest = pushmanifestparser.Manifest{}
		overrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = HandleDefaultRouteOverride(originalManifest, overrides)
	})

	When("the manifest has the no-route field", func() {
		BeforeEach(func() {
			originalManifest = pushmanifestparser.Manifest{
				Applications: []pushmanifestparser.Application{{NoRoute: true}},
			}
		})
		It("does not add default route", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(transformedManifest).To(Equal(pushmanifestparser.Manifest{
				Applications: []pushmanifestparser.Application{{NoRoute: true}},
			}))
		})

	})

	When("the manifest has the random-route field", func() {
		BeforeEach(func() {
			originalManifest = pushmanifestparser.Manifest{
				Applications: []pushmanifestparser.Application{{RandomRoute: true}},
			}
		})
		It("does not add default route", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(transformedManifest).To(Equal(pushmanifestparser.Manifest{
				Applications: []pushmanifestparser.Application{{RandomRoute: true}},
			}))

		})

	})

	// CLI doesnt know about the routes field but CAPI ignores defualt route if routes is specified
	// so we are ok adding defualt route even with the presence of a routes field

	When("the manifest has no routing fields", func() {
		BeforeEach(func() {
			originalManifest = pushmanifestparser.Manifest{
				Applications: []pushmanifestparser.Application{{}},
			}
		})
		It("does add default route", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(transformedManifest).To(Equal(pushmanifestparser.Manifest{
				Applications: []pushmanifestparser.Application{{DefaultRoute: true}},
			}))

		})

	})

})
