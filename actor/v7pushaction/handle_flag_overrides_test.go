package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleFlagOverrides", func() {
	var (
		pushActor           *Actor
		baseManifest        pushmanifestparser.Manifest
		flagOverrides       FlagOverrides
		transformedManifest pushmanifestparser.Manifest
		executeErr          error

		testFuncCallCount int
	)

	testTransformManifestFunc := func(manifest pushmanifestparser.Manifest, overrides FlagOverrides) (pushmanifestparser.Manifest, error) {
		testFuncCallCount += 1
		return manifest, nil
	}

	BeforeEach(func() {
		baseManifest = pushmanifestparser.Manifest{}
		flagOverrides = FlagOverrides{}
		testFuncCallCount = 0

		pushActor, _, _ = getTestPushActor()
		pushActor.TransformManifestSequence = []HandleFlagOverrideFunc{
			testTransformManifestFunc,
		}
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = pushActor.HandleFlagOverrides(baseManifest, flagOverrides)
	})

	It("calls each transform-manifest function", func() {
		Expect(testFuncCallCount).To(Equal(1))
		Expect(executeErr).NotTo(HaveOccurred())
		Expect(transformedManifest).To(Equal(baseManifest))
	})
})
