package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stack Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetStack", func() {
		Context("when the CC API client does not return any errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetStackReturns(
					ccv2.Stack{
						Name:        "some-stack",
						Description: "some stack description",
					},
					ccv2.Warnings{"get-stack-warning"},
					nil,
				)
			})

			It("returns the stack and all warnings", func() {
				stack, warnings, err := actor.GetStack("stack-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-stack-warning"))
				Expect(stack).To(Equal(Stack{
					Name:        "some-stack",
					Description: "some stack description",
				}))
			})
		})

		Context("when the stack does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetStackReturns(
					ccv2.Stack{},
					nil,
					ccerror.ResourceNotFoundError{},
				)
			})

			It("returns a StackNotFoundError", func() {
				_, _, err := actor.GetStack("stack-guid")
				Expect(err).To(MatchError(actionerror.StackNotFoundError{GUID: "stack-guid"}))
			})
		})

		Context("when the CC API client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get-stack-error")
				fakeCloudControllerClient.GetStackReturns(
					ccv2.Stack{},
					ccv2.Warnings{"stack-warning"},
					expectedErr,
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetStack("stack-guid")
				Expect(err).To(MatchError("get-stack-error"))
				Expect(warnings).To(ConsistOf("stack-warning"))
			})
		})
	})

	Describe("GetStackByName", func() {
		Context("when the CC API client does not return any errors", func() {
			Context("when it returns one stack", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetStacksReturns(
						[]ccv2.Stack{{
							Name:        "some-stack",
							Description: "some stack description",
						}},
						ccv2.Warnings{"get-stacks-warning"},
						nil,
					)
				})

				It("returns the stack and all warnings", func() {
					stack, warnings, err := actor.GetStackByName("some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-stacks-warning"))
					Expect(stack).To(Equal(Stack{
						Name:        "some-stack",
						Description: "some stack description",
					}))

					Expect(fakeCloudControllerClient.GetStacksCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetStacksArgsForCall(0)).To(Equal([]ccv2.QQuery{
						{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-stack"},
						},
					}))
				})
			})

			Context("when it returns no stacks", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetStacksReturns(
						[]ccv2.Stack{},
						ccv2.Warnings{"get-stacks-warning"},
						nil,
					)
				})

				It("returns a StackNotFoundError", func() {
					_, warnings, err := actor.GetStackByName("some-stack")
					Expect(err).To(MatchError(actionerror.StackNotFoundError{Name: "some-stack"}))
					Expect(warnings).To(ConsistOf("get-stacks-warning"))
				})
			})
		})

		Context("when the stack does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetStackReturns(
					ccv2.Stack{},
					nil,
					ccerror.ResourceNotFoundError{},
				)
			})

			It("returns a StackNotFoundError", func() {
				_, _, err := actor.GetStack("stack-guid")
				Expect(err).To(MatchError(actionerror.StackNotFoundError{GUID: "stack-guid"}))
			})
		})

		Context("when the CC API client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get-stack-error")
				fakeCloudControllerClient.GetStackReturns(
					ccv2.Stack{},
					ccv2.Warnings{"stack-warning"},
					expectedErr,
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetStack("stack-guid")
				Expect(err).To(MatchError("get-stack-error"))
				Expect(warnings).To(ConsistOf("stack-warning"))
			})
		})
	})
})
