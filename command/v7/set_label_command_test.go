package v7_test

import (
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("set-label command", func() {
	var (
		cmd             SetLabelCommand
		resourceName    string
		fakeLabelSetter *v7fakes.FakeLabelSetter

		executeErr error
	)

	Context("shared validations", func() {
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
					Labels:       []string{"FOO=BAR", "ENV=FAKE"},
				}
				cmd.BuildpackStack = "some-stack"

			})

			It("calls execute with the right parameters", func() {
				executeErr = cmd.Execute(nil)

				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeLabelSetter.ExecuteCallCount()).To(Equal(1))
				targetResource, labels := fakeLabelSetter.ExecuteArgsForCall(0)
				Expect(targetResource.ResourceType).To(Equal(cmd.RequiredArgs.ResourceType))
				Expect(targetResource.ResourceName).To(Equal(cmd.RequiredArgs.ResourceName))
				Expect(targetResource.BuildpackStack).To(Equal(cmd.BuildpackStack))
				Expect(labels).To(Equal(map[string]types.NullString{
					"FOO": types.NewNullString("BAR"),
					"ENV": types.NewNullString("FAKE"),
				}))
			})
		})

	})
})
