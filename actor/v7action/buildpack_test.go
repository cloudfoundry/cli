package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Buildpack", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _ = NewTestActor()
	})

	Describe("GetBuildpacks", func() {
		var (
			buildpacks []Buildpack
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			buildpacks, warnings, executeErr = actor.GetBuildpacks()
		})

		When("getting buildpacks fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetBuildpacksReturns(
					nil,
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("some-error"))
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
			})
		})

		When("getting buildpacks is successful", func() {
			BeforeEach(func() {
				ccBuildpacks := []ccv3.Buildpack{
					{Name: "buildpack-1", Position: 1},
					{Name: "buildpack-2", Position: 2},
				}

				fakeCloudControllerClient.GetBuildpacksReturns(
					ccBuildpacks,
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					nil)
			})

			It("returns the buildpacks and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
				Expect(buildpacks).To(Equal([]Buildpack{
					{Name: "buildpack-1", Position: 1},
					{Name: "buildpack-2", Position: 2},
				}))

				Expect(fakeCloudControllerClient.GetBuildpacksCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetBuildpacksArgsForCall(0)).To(ConsistOf(ccv3.Query{
					Key:    ccv3.OrderBy,
					Values: []string{ccv3.PositionOrder},
				}))
			})
		})
	})
})
