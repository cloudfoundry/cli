package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/command/flag"
	v7 "code.cloudfoundry.org/cli/v8/command/v7"
	"code.cloudfoundry.org/cli/v8/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v8/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("unset-label command", func() {
	var (
		cmd             v7.UnsetLabelCommand
		resourceName    string
		fakeLabelSetter *v7fakes.FakeLabelUnsetter

		executeErr error
	)

	BeforeEach(func() {
		fakeLabelSetter = new(v7fakes.FakeLabelUnsetter)
		cmd = v7.UnsetLabelCommand{
			LabelUnsetter: fakeLabelSetter,
		}

		cmd.RequiredArgs = flag.UnsetLabelArgs{
			ResourceType: "anything",
			ResourceName: resourceName,
			LabelKeys:    []string{"FOO", "ENV"},
		}
		cmd.BuildpackStack = "some-stack"
		cmd.ServiceBroker = "some-service-broker"
		cmd.ServiceOffering = "some-service-offering"
	})

	It("calls execute with the right parameters", func() {
		executeErr = cmd.Execute(nil)

		Expect(executeErr).ToNot(HaveOccurred())
		Expect(fakeLabelSetter.ExecuteCallCount()).To(Equal(1))
		targetResource, keys := fakeLabelSetter.ExecuteArgsForCall(0)
		Expect(targetResource.ResourceType).To(Equal(cmd.RequiredArgs.ResourceType))
		Expect(targetResource.ResourceName).To(Equal(cmd.RequiredArgs.ResourceName))
		Expect(targetResource.BuildpackStack).To(Equal(cmd.BuildpackStack))
		Expect(targetResource.ServiceBroker).To(Equal(cmd.ServiceBroker))
		Expect(targetResource.ServiceOffering).To(Equal(cmd.ServiceOffering))
		Expect(keys).To(Equal(map[string]types.NullString{
			"FOO": types.NewNullString(),
			"ENV": types.NewNullString(),
		}))
	})

	When("--source flag is provided", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.UnsetLabelArgs{
				ResourceType: "route-policy",
				ResourceName: "some-route",
				LabelKeys:    []string{"FOO"},
			}
			cmd.RoutePolicySource = "cf:app:some-app-guid"
		})

		It("passes the source to the label unsetter", func() {
			executeErr = cmd.Execute(nil)

			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeLabelSetter.ExecuteCallCount()).To(Equal(1))
			targetResource, _ := fakeLabelSetter.ExecuteArgsForCall(0)
			Expect(targetResource.RoutePolicySource).To(Equal("cf:app:some-app-guid"))
		})
	})

	When("--source flag is provided with a non-route-policy resource", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.UnsetLabelArgs{
				ResourceType: "app",
				ResourceName: "my-app",
				LabelKeys:    []string{"FOO"},
			}
			cmd.RoutePolicySource = "cf:app:some-app-guid"
			fakeLabelSetter.ExecuteReturns(errors.New("argument combination error"))
		})

		It("returns an error", func() {
			executeErr = cmd.Execute(nil)

			Expect(executeErr).To(HaveOccurred())
		})
	})
})
