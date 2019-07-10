package v7pushaction_test

import (
	"fmt"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/util/manifestparser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreatePushPlans", func() {
	var (
		pushActor *Actor

		appNameArg         string
		spaceGUID          string
		orgGUID            string
		fakeManifestParser *v7pushactionfakes.FakeManifestParser
		flagOverrides      FlagOverrides

		pushPlans  []PushPlan
		executeErr error

		testUpdatePlanCount int
	)

	testUpdatePlan := func(pushState PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
		testUpdatePlanCount += 1
		pushState.Application.Name = manifestApp.Name
		return pushState, nil
	}

	BeforeEach(func() {
		pushActor, _, _ = getTestPushActor()
		pushActor.PreparePushPlanSequence = []UpdatePushPlanFunc{testUpdatePlan, testUpdatePlan}

		appNameArg = "my-app"
		orgGUID = "org"
		spaceGUID = "space"
		flagOverrides = FlagOverrides{}
		fakeManifestParser = new(v7pushactionfakes.FakeManifestParser)

		testUpdatePlanCount = 0
	})

	JustBeforeEach(func() {
		pushPlans, executeErr = pushActor.CreatePushPlans(appNameArg, spaceGUID, orgGUID, fakeManifestParser, flagOverrides)
	})

	AssertNoExecuteErr := func() {
		It("returns nil", func() {
			Expect(executeErr).ToNot(HaveOccurred())
		})
	}

	AssertPushPlanLength := func(length int) {
		It(fmt.Sprintf("creates a []pushPlan with length %d", length), func() {
			Expect(pushPlans).To(HaveLen(length))
		})
	}

	Describe("Manifest", func() {
		BeforeEach(func() {
			fakeManifestParser.AppsReturns([]manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Name: "my-app",
					},
					FullUnmarshalledApplication: nil,
				},
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Name: "spencers-app",
					},
					FullUnmarshalledApplication: nil,
				},
			})

			fakeManifestParser.ContainsManifestReturns(true)

			appNameArg = ""
		})

		AssertNoExecuteErr()
		AssertPushPlanLength(2)

		It("it creates pushPlans based on the apps in the manifest", func() {
			Expect(pushPlans[0].Application.Name).To(Equal("my-app"))
			Expect(pushPlans[1].Application.Name).To(Equal("spencers-app"))
		})
	})

	Describe("Org and Space GUID", func() {
		It("creates pushPlans with org and space GUIDs", func() {
			Expect(pushPlans[0].SpaceGUID).To(Equal(spaceGUID))
			Expect(pushPlans[0].OrgGUID).To(Equal(orgGUID))
		})
	})

	Describe("update push plans", func() {
		It("runs through all the update push plan functions", func() {
			Expect(testUpdatePlanCount).To(Equal(2))
		})
	})
})
