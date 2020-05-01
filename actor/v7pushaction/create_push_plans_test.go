package v7pushaction_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/util/manifestparser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreatePushPlans", func() {
	var (
		pushActor   *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		manifest      manifestparser.Manifest
		spaceGUID     string
		orgGUID       string
		flagOverrides FlagOverrides

		pushPlans  []PushPlan
		executeErr error
		warnings   v7action.Warnings

		testUpdatePlanCount int
	)

	testUpdatePlan := func(pushState PushPlan, overrides FlagOverrides) (PushPlan, error) {
		testUpdatePlanCount += 1
		return pushState, nil
	}

	BeforeEach(func() {
		pushActor, fakeV7Actor, _ = getTestPushActor()
		pushActor.PreparePushPlanSequence = []UpdatePushPlanFunc{testUpdatePlan, testUpdatePlan}

		manifest = manifestparser.Manifest{
			Applications: []manifestparser.Application{
				{Name: "name-1", Path: "path1"},
				{Name: "name-2", Path: "path2", Docker: &manifestparser.Docker{Image: "image", Username: "uname"}},
			},
		}
		orgGUID = "org"
		spaceGUID = "space"
		flagOverrides = FlagOverrides{
			DockerPassword: "passwd",
		}

		testUpdatePlanCount = 0
	})

	JustBeforeEach(func() {
		pushPlans, warnings, executeErr = pushActor.CreatePushPlans(spaceGUID, orgGUID, manifest, flagOverrides)
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

	It("delegates to the V7actor to gets the apps", func() {
		Expect(fakeV7Actor.GetApplicationsByNamesAndSpaceCallCount()).To(Equal(1))

		actualAppNames, actualSpaceGUID := fakeV7Actor.GetApplicationsByNamesAndSpaceArgsForCall(0)
		Expect(actualAppNames).To(ConsistOf("name-1", "name-2"))
		Expect(actualSpaceGUID).To(Equal(spaceGUID))
	})

	When("getting the apps fails", func() {
		BeforeEach(func() {
			fakeV7Actor.GetApplicationsByNamesAndSpaceReturns(nil, v7action.Warnings{"get-apps-warning"}, errors.New("get-apps-error"))
		})

		It("returns errors and warnings", func() {
			Expect(executeErr).To(MatchError("get-apps-error"))
			Expect(warnings).To(ConsistOf("get-apps-warning"))
		})
	})

	When("getting the apps succeeds", func() {
		BeforeEach(func() {
			fakeV7Actor.GetApplicationsByNamesAndSpaceReturns(
				[]resources.Application{
					{Name: "name-1", GUID: "app-guid-1"},
					{Name: "name-2", GUID: "app-guid-2"},
				},
				v7action.Warnings{"get-apps-warning"},
				nil,
			)
		})
		It("runs through all the update push plan functions", func() {
			Expect(testUpdatePlanCount).To(Equal(4))
		})

		AssertNoExecuteErr()
		AssertPushPlanLength(2)

		It("returns warnings", func() {
			Expect(warnings).To(ConsistOf("get-apps-warning"))
		})

		It("it creates pushPlans based on the apps in the manifest", func() {
			Expect(pushPlans[0].Application.Name).To(Equal("name-1"))
			Expect(pushPlans[0].Application.GUID).To(Equal("app-guid-1"))
			Expect(pushPlans[0].SpaceGUID).To(Equal(spaceGUID))
			Expect(pushPlans[0].OrgGUID).To(Equal(orgGUID))
			Expect(pushPlans[0].DockerImageCredentials.Path).To(Equal(""))
			Expect(pushPlans[0].DockerImageCredentials.Username).To(Equal(""))
			Expect(pushPlans[0].DockerImageCredentials.Password).To(Equal(""))
			Expect(pushPlans[0].BitsPath).To(Equal("path1"))
			Expect(pushPlans[1].Application.Name).To(Equal("name-2"))
			Expect(pushPlans[1].Application.GUID).To(Equal("app-guid-2"))
			Expect(pushPlans[1].SpaceGUID).To(Equal(spaceGUID))
			Expect(pushPlans[1].OrgGUID).To(Equal(orgGUID))
			Expect(pushPlans[1].DockerImageCredentials.Path).To(Equal("image"))
			Expect(pushPlans[1].DockerImageCredentials.Username).To(Equal("uname"))
			Expect(pushPlans[1].DockerImageCredentials.Password).To(Equal("passwd"))
			Expect(pushPlans[1].BitsPath).To(Equal("path2"))
		})

	})
})
