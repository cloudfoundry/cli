package v7action_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource Matching", func() {
	var (
		resources                 []sharedaction.V3Resource
		executeErr                error
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		actor                     *Actor

		matchedResources []sharedaction.V3Resource

		warnings Warnings
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _, _ = NewTestActor()
		resources = []sharedaction.V3Resource{}
	})

	JustBeforeEach(func() {
		matchedResources, warnings, executeErr = actor.ResourceMatch(resources)
	})

	When("The cc client succeeds", func() {
		BeforeEach(func() {
			for i := 1; i <= constant.MaxNumberOfResourcesForMatching+1; i++ {
				resources = append(resources, sharedaction.V3Resource{
					FilePath:    fmt.Sprintf("path/to/file/%d", i),
					SizeInBytes: 1,
				})
			}
			resources = append(resources, sharedaction.V3Resource{
				FilePath:    "empty-file",
				SizeInBytes: 0,
			})

			fakeCloudControllerClient.ResourceMatchReturnsOnCall(0, []ccv3.Resource{{FilePath: "path/to/file"}}, ccv3.Warnings{"this-is-a-warning"}, nil)
			fakeCloudControllerClient.ResourceMatchReturnsOnCall(1, []ccv3.Resource{{FilePath: "path/to/other-file"}}, ccv3.Warnings{"this-is-another-warning"}, nil)
		})

		It("passes through the list of resources with no 0 length resources", func() {
			Expect(fakeCloudControllerClient.ResourceMatchCallCount()).To(Equal(2))

			passedResources := fakeCloudControllerClient.ResourceMatchArgsForCall(0)
			Expect(passedResources).To(HaveLen(constant.MaxNumberOfResourcesForMatching))
			Expect(passedResources[0].FilePath).To(MatchRegexp("path/to/file/\\d+"))

			passedResources = fakeCloudControllerClient.ResourceMatchArgsForCall(1)
			Expect(passedResources).To(HaveLen(1))
			Expect(passedResources[0].FilePath).To(MatchRegexp("path/to/file/\\d+"))
		})

		It("returns a list of sharedAction V3Resources and warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(warnings).To(ConsistOf("this-is-a-warning", "this-is-another-warning"))
			Expect(matchedResources).To(ConsistOf(
				sharedaction.V3Resource{FilePath: "path/to/file"},
				sharedaction.V3Resource{FilePath: "path/to/other-file"}))
		})
	})

	When("The cc client errors", func() {
		BeforeEach(func() {
			resources = []sharedaction.V3Resource{{SizeInBytes: 1}}
			fakeCloudControllerClient.ResourceMatchReturns(nil, ccv3.Warnings{"this-is-a-warning"}, errors.New("boom"))
		})

		It("raises the error", func() {
			Expect(executeErr).To(MatchError("boom"))
			Expect(warnings).To(ConsistOf(ccv3.Warnings{"this-is-a-warning"}))
		})
	})
})
