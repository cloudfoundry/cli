package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/command/flag"
	. "code.cloudfoundry.org/cli/v9/command/v7"
	"code.cloudfoundry.org/cli/v9/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v9/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("set-label command", func() {
	var (
		cmd             SetLabelCommand
		resourceName    string
		fakeLabelSetter *v7fakes.FakeLabelSetter

		executeErr error
	)

	BeforeEach(func() {
		fakeLabelSetter = new(v7fakes.FakeLabelSetter)
		cmd = SetLabelCommand{
			LabelSetter: fakeLabelSetter,
		}
	})

	When("some provided labels do not have a value part", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "anything",
				ResourceName: resourceName,
				Labels:       []string{"FOO=BAR", "MISSING_EQUALS", "ENV=FAKE"},
			}
		})

		It("complains about the missing equal sign", func() {
			err := cmd.Execute(nil)
			Expect(err).To(MatchError("Metadata error: no value provided for label 'MISSING_EQUALS'"))
			Expect(err).To(HaveOccurred())
		})
	})

	When("all the provided labels are valid", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "anything",
				ResourceName: resourceName,
				Labels:       []string{"FOO=BAZ", "FOO=BAR", "ENV=FAKE"},
			}
			cmd.BuildpackStack = "some-stack"
			cmd.ServiceBroker = "some-service-broker"
			cmd.ServiceOffering = "some-service-offering"
		})

		It("calls execute with the right parameters", func() {
			executeErr = cmd.Execute(nil)

			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeLabelSetter.ExecuteCallCount()).To(Equal(1))
			targetResource, labels := fakeLabelSetter.ExecuteArgsForCall(0)
			Expect(targetResource.ResourceType).To(Equal(cmd.RequiredArgs.ResourceType))
			Expect(targetResource.ResourceName).To(Equal(cmd.RequiredArgs.ResourceName))
			Expect(targetResource.BuildpackStack).To(Equal(cmd.BuildpackStack))
			Expect(targetResource.ServiceBroker).To(Equal(cmd.ServiceBroker))
			Expect(targetResource.ServiceOffering).To(Equal(cmd.ServiceOffering))
			Expect(labels).To(Equal(map[string]types.NullString{
				"FOO": types.NewNullString("BAR"),
				"ENV": types.NewNullString("FAKE"),
			}))
		})
	})

	When("setting labels on a route-policy resource", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "route-policy",
				ResourceName: "foo.example.com/the-path",
				Labels:       []string{"FOO=BAR"},
			}
		})

		It("passes the route URL as ResourceName to the label setter without parsing", func() {
			executeErr = cmd.Execute(nil)

			Expect(executeErr).ToNot(HaveOccurred())
			targetResource, _ := fakeLabelSetter.ExecuteArgsForCall(0)
			Expect(targetResource.ResourceType).To(Equal("route-policy"))
			Expect(targetResource.ResourceName).To(Equal("foo.example.com/the-path"))
		})
	})

	When("--source flag is provided", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "route-policy",
				ResourceName: "some-route",
				Labels:       []string{"FOO=BAR"},
			}
			cmd.RoutePolicySource = "cf:app:some-app-guid"
		})

		It("passes the source to the label setter", func() {
			executeErr = cmd.Execute(nil)

			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeLabelSetter.ExecuteCallCount()).To(Equal(1))
			targetResource, _ := fakeLabelSetter.ExecuteArgsForCall(0)
			Expect(targetResource.RoutePolicySource).To(Equal("cf:app:some-app-guid"))
		})
	})

	When("--source flag is provided with a non-route-policy resource", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.SetLabelArgs{
				ResourceType: "app",
				ResourceName: "my-app",
				Labels:       []string{"FOO=BAR"},
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
