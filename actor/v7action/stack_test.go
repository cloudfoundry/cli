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

var _ = Describe("Stack", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeConfig := new(v7actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil)
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

	Describe("GetStacks", func() {
		var (
			ccv3Stacks []ccv3.Stack
			stacks     []Stack

			stack1Name        string
			stack1Description string
			stack2Name        string
			stack2Description string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			ccv3Stacks = []ccv3.Stack{
				{Name: stack1Name, Description: stack1Description},
				{Name: stack2Name, Description: stack2Description},
			}
		})

		JustBeforeEach(func() {
			stacks, warnings, executeErr = actor.GetStacks()
		})

		When("getting stacks returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeCloudControllerClient.GetStacksReturns(
					[]ccv3.Stack{},
					ccv3.Warnings{"warning-1", "warning-2"}, expectedErr)
			})

			It("returns warnings and the error", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		When("the GetStacks call is successful", func() {
			When("the cloud controller returns back stacks", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetStacksReturns(
						ccv3Stacks,
						ccv3.Warnings{"some-stack-warning"}, nil)
				})

				It("returns back the stacks and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(stacks).To(ConsistOf(Stack{Name: stack1Name, Description: stack1Description}, Stack{Name: stack2Name, Description: stack2Description}))
					Expect(warnings).To(ConsistOf("some-stack-warning"))
					Expect(fakeCloudControllerClient.GetStacksCallCount()).To(Equal(1))
				})
			})

			When("the GetStacks call is unsuccessful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetStacksReturns(
						nil,
						ccv3.Warnings{"some-stack-warning"},
						errors.New("some-error"))
				})

				It("returns an error and warnings", func() {
					Expect(executeErr).To(MatchError("some-error"))
					Expect(warnings).To(ConsistOf("some-stack-warning"))
				})
			})
		})
	})
})
