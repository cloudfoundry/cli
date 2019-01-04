package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-service Command", func() {
	var (
		cmd             CreateServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeCreateServiceActor
		executeErr      error
		extraArgs       []string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeCreateServiceActor)

		extraArgs = nil
		cmd = CreateServiceCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			RequiredArgs: flag.CreateServiceArgs{
				Service:         "cool-broker",
				ServiceInstance: "cool-service",
				ServicePlan:     "cool-plan",
			},
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(extraArgs)
	})

	When("the user provides extra arguments", func() {
		BeforeEach(func() {
			extraArgs = []string{"some-extra-arg"}
		})

		It("fails with a TooManyArgumentsError", func() {
			Expect(executeErr).To(MatchError(translatableerror.TooManyArgumentsError{
				ExtraArgument: "some-extra-arg",
			}))
		})
	})

	It("checks the user is logged in, and targeting an org and space", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		orgChecked, spaceChecked := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(orgChecked).To(BeTrue())
		Expect(spaceChecked).To(BeTrue())
	})

	When("checking the target succeeds", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(nil)
			fakeConfig.CurrentUserReturns(
				configv3.User{Name: "some-user"},
				nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
		})

		When("creating a service is successful", func() {
			BeforeEach(func() {
				fakeActor.CreateServiceInstanceReturns(v2action.ServiceInstance{LastOperation: ccv2.LastOperation{State: constant.LastOperationSucceeded}}, []string{"a-warning", "another-warning"}, nil)
			})

			It("displays a success message indicating that it is creating the service", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say("Creating service instance cool-service in org some-org / space some-space as some-user\\.\\.\\."))
				Expect(testUI.Out).To(Say("OK"))
			})

			It("passes the correct args when creating the service instance", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeActor.CreateServiceInstanceCallCount()).To(Equal(1))
				spaceGUID, service, servicePlan, serviceInstance, _, _ := fakeActor.CreateServiceInstanceArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(service).To(Equal("cool-broker"))
				Expect(servicePlan).To(Equal("cool-plan"))
				Expect(serviceInstance).To(Equal("cool-service"))
			})

			It("displays all warnings", func() {
				Expect(testUI.Err).To(Say("a-warning"))
				Expect(testUI.Err).To(Say("another-warning"))
			})

			Context("the user passes in tags", func() {
				BeforeEach(func() {
					cmd.Tags = []string{"tag-1", "tag-2"}
				})

				It("passes the tags as args when creating the service instance", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(fakeActor.CreateServiceInstanceCallCount()).To(Equal(1))
					_, _, _, _, _, tags := fakeActor.CreateServiceInstanceArgsForCall(0)
					Expect(tags).To(Equal([]string{"tag-1", "tag-2"}))
				})
			})

			Context("the user passes in parameters", func() {
				BeforeEach(func() {
					cmd.ParametersAsJSON = map[string]interface{}{
						"some-key": "some-value",
					}
				})

				It("passes the parameters as args when creating the service instance", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(fakeActor.CreateServiceInstanceCallCount()).To(Equal(1))
					_, _, _, _, params, _ := fakeActor.CreateServiceInstanceArgsForCall(0)
					Expect(params).To(Equal(map[string]interface{}{"some-key": "some-value"}))
				})
			})

			When("the create is in progress", func() {
				BeforeEach(func() {
					fakeActor.CreateServiceInstanceReturns(v2action.ServiceInstance{LastOperation: ccv2.LastOperation{State: constant.LastOperationInProgress}}, []string{"a-warning", "another-warning"}, nil)
				})

				It("displays a message indicating that create service is in progress", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(fakeActor.CreateServiceInstanceCallCount()).To(Equal(1))
					Expect(testUI.Out).To(Say("Creating service instance cool-service in org some-org / space some-space as some-user\\.\\.\\."))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("Create in progress\\. Use 'cf services' or 'cf service cool-service' to check operation status\\."))
				})
			})
		})

		When("creating a service returns an error", func() {
			BeforeEach(func() {
				fakeActor.CreateServiceInstanceReturns(v2action.ServiceInstance{}, []string{"a-warning", "another-warning"}, errors.New("explode"))
			})

			It("returns the error and displays all warnings", func() {
				Expect(executeErr).To(MatchError("explode"))
				Expect(fakeActor.CreateServiceInstanceCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("Creating service instance cool-service in org some-org / space some-space as some-user\\.\\.\\."))
				Expect(testUI.Err).To(Say("a-warning"))
				Expect(testUI.Err).To(Say("another-warning"))
			})

			When("the service instance name is taken", func() {
				BeforeEach(func() {
					fakeActor.CreateServiceInstanceReturns(v2action.ServiceInstance{}, []string{"a-warning", "another-warning"}, ccerror.ServiceInstanceNameTakenError{})
				})

				It("succeeds, displaying warnings, 'OK' and an informative message", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(testUI.Err).To(Say("a-warning"))
					Expect(testUI.Err).To(Say("another-warning"))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("Service cool-service already exists"))
				})
			})
		})

		When("fetching the current user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("explode"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("explode"))
			})
		})
	})

	When("checking the target returns an error", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("explode"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("explode"))
		})
	})
})
