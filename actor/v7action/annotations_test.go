package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("annotations", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		fakeSharedActor           *v7actionfakes.FakeSharedActor
		fakeConfig                *v7actionfakes.FakeConfig
		warnings                  Warnings
		executeErr                error
		annotations               map[string]types.NullString
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeSharedActor = new(v7actionfakes.FakeSharedActor)
		fakeConfig = new(v7actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, nil, nil, nil)
	})

	Describe("GetRevisionAnnotations", func() {
		JustBeforeEach(func() {
			annotations, warnings, executeErr = actor.GetRevisionAnnotations("some-guid")
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRevisionReturns(
					resources.Revision{GUID: "some-guid"},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
			})

			It("gets the revision", func() {
				Expect(fakeCloudControllerClient.GetRevisionCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRevisionArgsForCall(0)).To(Equal("some-guid"))
			})

			When("there are no annotations on a revision", func() {
				It("returns an empty map", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(annotations).To(BeEmpty())
				})
			})

			When("there are annotations", func() {
				var expectedAnnotations map[string]types.NullString

				BeforeEach(func() {
					expectedAnnotations = map[string]types.NullString{"key1": types.NewNullString("value1"), "key2": types.NewNullString("value2")}
					fakeCloudControllerClient.GetRevisionReturns(
						resources.Revision{
							GUID: "some-guid",
							Metadata: &resources.Metadata{
								Annotations: expectedAnnotations,
							},
						},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
				})
				It("returns the annotations", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(annotations).To(Equal(expectedAnnotations))
				})
			})
		})

		When("there is a client error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRevisionReturns(
					resources.Revision{GUID: "some-guid"},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					errors.New("get-revision-error"),
				)
			})
			When("GetRevision fails", func() {
				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(executeErr).To(MatchError("get-revision-error"))
				})
			})
		})
	})
})
