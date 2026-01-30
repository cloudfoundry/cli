package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	. "code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
	. "github.com/onsi/ginkgo/v2"
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
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil, nil, nil)
	})

	Describe("Get stack by name", func() {
		var expectedErr error
		var err error
		var warnings Warnings
		var stack resources.Stack

		JustBeforeEach(func() {
			stack, warnings, err = actor.GetStackByName("some-stack-name")
		})

		Describe("there are errors", func() {
			When("the client errors", func() {
				BeforeEach(func() {
					expectedErr = errors.New("CC Error")
					fakeCloudControllerClient.GetStacksReturns(
						[]resources.Stack{},
						ccv3.Warnings{"warning-1", "warning-2"},
						expectedErr,
					)
				})

				It("returns the same error", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			When("the stack does not exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetStacksReturns(
						[]resources.Stack{},
						ccv3.Warnings{"warning-1", "warning-2"},
						actionerror.StackNotFoundError{Name: "some-stack-name"},
					)
				})

				It("returns a StackNotFound error", func() {
					Expect(err).To(MatchError(actionerror.StackNotFoundError{Name: "some-stack-name"}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})

		Context("there are no errors", func() {

			When("the stack exists without state", func() {
				expectedStack := resources.Stack{
					GUID:        "some-stack-guid",
					Name:        "some-stack-name",
					Description: "Some stack desc",
				}

				expectedParams := []ccv3.Query{
					{Key: ccv3.NameFilter, Values: []string{"some-stack-name"}},
					{Key: ccv3.PerPage, Values: []string{"1"}},
					{Key: ccv3.Page, Values: []string{"1"}},
				}

				BeforeEach(func() {
					fakeCloudControllerClient.GetStacksReturns(
						[]resources.Stack{expectedStack},
						ccv3.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("returns the desired stack", func() {

					actualParams := fakeCloudControllerClient.GetStacksArgsForCall(0)
					Expect(actualParams).To(Equal(expectedParams))
					Expect(fakeCloudControllerClient.GetStacksCallCount()).To(Equal(1))
					Expect(stack.GUID).To(Equal(expectedStack.GUID))
					Expect(stack.Name).To(Equal(expectedStack.Name))
					Expect(stack.Description).To(Equal(expectedStack.Description))
					Expect(stack.State).To(Equal(""))
					Expect(err).To(BeNil())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			When("the stack exists with state", func() {
				expectedStack := resources.Stack{
					GUID:        "some-stack-guid",
					Name:        "some-stack-name",
					Description: "Some stack desc",
					State:       "ACTIVE",
				}

				expectedParams := []ccv3.Query{
					{Key: ccv3.NameFilter, Values: []string{"some-stack-name"}},
					{Key: ccv3.PerPage, Values: []string{"1"}},
					{Key: ccv3.Page, Values: []string{"1"}},
				}

				BeforeEach(func() {
					fakeCloudControllerClient.GetStacksReturns(
						[]resources.Stack{expectedStack},
						ccv3.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("returns the desired stack with state", func() {

					actualParams := fakeCloudControllerClient.GetStacksArgsForCall(0)
					Expect(actualParams).To(Equal(expectedParams))
					Expect(fakeCloudControllerClient.GetStacksCallCount()).To(Equal(1))
					Expect(stack.GUID).To(Equal(expectedStack.GUID))
					Expect(stack.Name).To(Equal(expectedStack.Name))
					Expect(stack.Description).To(Equal(expectedStack.Description))
					Expect(stack.State).To(Equal(resources.StackStateActive))
					Expect(err).To(BeNil())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})
	})

	Describe("GetStacks", func() {
		var (
			apiReturnedStacks []resources.Stack
			stacks            []resources.Stack

			stack1Name        string
			stack1Description string
			stack2Name        string
			stack2Description string

			warnings      Warnings
			executeErr    error
			labelSelector string
		)

		BeforeEach(func() {
			apiReturnedStacks = []resources.Stack{
				{Name: stack1Name, Description: stack1Description},
				{Name: stack2Name, Description: stack2Description},
			}
		})

		JustBeforeEach(func() {
			stacks, warnings, executeErr = actor.GetStacks(labelSelector)
		})

		When("getting stacks returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeCloudControllerClient.GetStacksReturns(
					[]resources.Stack{},
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
						apiReturnedStacks,
						ccv3.Warnings{"some-stack-warning"}, nil)
				})

				It("returns back the stacks and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(stacks).To(ConsistOf(resources.Stack{Name: stack1Name, Description: stack1Description}, resources.Stack{Name: stack2Name, Description: stack2Description}))
					Expect(warnings).To(ConsistOf("some-stack-warning"))
					Expect(fakeCloudControllerClient.GetStacksCallCount()).To(Equal(1))
				})
				When("a label selector is passed in", func() {
					BeforeEach(func() {
						labelSelector = "some-label-selector"
					})

					It("passes the selector to API", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeCloudControllerClient.GetStacksCallCount()).To(Equal(1))

						expectedQuery := []ccv3.Query{
							{Key: ccv3.LabelSelectorFilter, Values: []string{"some-label-selector"}},
						}
						actualQuery := fakeCloudControllerClient.GetStacksArgsForCall(0)
						Expect(actualQuery).To(Equal(expectedQuery))

					})
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

	Describe("UpdateStack", func() {
		var (
			stackGUID  string
			state      string
			stack      resources.Stack
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			stackGUID = "some-stack-guid"
			state = "DEPRECATED"
		})

		JustBeforeEach(func() {
			stack, warnings, executeErr = actor.UpdateStack(stackGUID, state)
		})

		When("the cloud controller request is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateStackReturns(
					resources.Stack{
						GUID:        "some-stack-guid",
						Name:        "some-stack",
						Description: "some description",
						State:       "DEPRECATED",
					},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns the updated stack and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(stack).To(Equal(resources.Stack{
					GUID:        "some-stack-guid",
					Name:        "some-stack",
					Description: "some description",
					State:       "DEPRECATED",
				}))

				Expect(fakeCloudControllerClient.UpdateStackCallCount()).To(Equal(1))
				actualGUID, actualState := fakeCloudControllerClient.UpdateStackArgsForCall(0)
				Expect(actualGUID).To(Equal(stackGUID))
				Expect(actualState).To(Equal(state))
			})
		})

		When("the cloud controller request fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateStackReturns(
					resources.Stack{},
					ccv3.Warnings{"warning-1"},
					errors.New("some-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})
})
