package v7action_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stack Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("Get stack by name", func() {

		var expectedErr error
		var err error
		var warnings Warnings
		var stack Stack

		JustBeforeEach(func() {
			stack, warnings, err = actor.GetStackByName("some-stack-name")
		})

		Describe("When there are errors", func() {
			When("The client errors", func() {
				BeforeEach(func() {
					expectedErr = errors.New("CC Error")
					fakeCloudControllerClient.GetStacksReturns(
						[]ccv3.Stack{},
						ccv3.Warnings{"warning-1", "warning-2"},
						expectedErr,
					)
				})

				It("Returns the same error", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			When("The stack does not exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetStacksReturns(
						[]ccv3.Stack{},
						ccv3.Warnings{"warning-1", "warning-2"},
						actionerror.StackNotFoundError{Name: "some-stack-name"},
					)
				})

				It("Returns a StackNotFound error", func() {
					Expect(err).To(MatchError(actionerror.StackNotFoundError{Name: "some-stack-name"}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})

		Context("When there are no errors", func() {

			When("The stack exists", func() {
				expectedStack := ccv3.Stack{
					GUID:        "some-stack-guid",
					Name:        "some-stack-name",
					Description: "Some stack desc",
				}

				expectedParams := []ccv3.Query{{Key: ccv3.NameFilter, Values: []string{"some-stack-name"}}}

				BeforeEach(func() {
					fakeCloudControllerClient.GetStacksReturns(
						[]ccv3.Stack{expectedStack},
						ccv3.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("Returns the desired stack", func() {

					actualParams := fakeCloudControllerClient.GetStacksArgsForCall(0)
					Expect(actualParams).To(Equal(expectedParams))
					Expect(fakeCloudControllerClient.GetStacksCallCount()).To(Equal(1))
					Expect(stack.GUID).To(Equal(expectedStack.GUID))
					Expect(stack.Name).To(Equal(expectedStack.Name))
					Expect(stack.Description).To(Equal(expectedStack.Description))
					Expect(err).To(BeNil())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})

	})
})
