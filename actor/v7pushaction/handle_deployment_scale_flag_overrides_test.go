package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/v9/actor/v7pushaction"
	"code.cloudfoundry.org/cli/v9/util/manifestparser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleDeploymentScaleFlagOverrides", func() {
	var (
		pushActor           *Actor
		baseManifest        manifestparser.Manifest
		flagOverrides       FlagOverrides
		transformedManifest manifestparser.Manifest
		executeErr          error

		testFuncCallCount int
	)

	testTransformManifestFunc := func(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
		testFuncCallCount += 1
		return manifest, nil
	}

	BeforeEach(func() {
		baseManifest = manifestparser.Manifest{}
		flagOverrides = FlagOverrides{}
		testFuncCallCount = 0

		pushActor, _, _ = getTestPushActor()
		pushActor.TransformManifestSequenceForDeployment = []HandleFlagOverrideFunc{
			testTransformManifestFunc,
		}
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = pushActor.HandleDeploymentScaleFlagOverrides(baseManifest, flagOverrides)
	})

	It("calls each transform-manifest-for-deployment function", func() {
		Expect(testFuncCallCount).To(Equal(1))
		Expect(executeErr).NotTo(HaveOccurred())
		Expect(transformedManifest).To(Equal(baseManifest))
	})
})
