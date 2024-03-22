package v7_test

import (
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("ConfigCommand", func() {
	var (
		cmd        v7.ConfigCommand
		testUI     *ui.UI
		fakeConfig *commandfakes.FakeConfig
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)

		cmd = ConfigCommand{
			UI:     testUI,
			Config: fakeConfig,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("there are no flags given", func() {
		It("returns an error", func() {
			Expect(executeErr).To(MatchError(
				translatableerror.IncorrectUsageError{Message: "at least one flag must be provided"},
			))
		})
	})

	When("using the async timeout flag", func() {
		BeforeEach(func() {
			cmd.AsyncTimeout = flag.Timeout{NullInt: types.NullInt{IsSet: true, Value: 2}}
		})

		It("successfully updates the config", func() {
			Expect(executeErr).To(Not(HaveOccurred()))
			Expect(fakeConfig.SetAsyncTimeoutCallCount()).To(Equal(1))
			value := fakeConfig.SetAsyncTimeoutArgsForCall(0)
			Expect(value).To(Equal(2))
		})
	})

	When("using the color flag", func() {
		BeforeEach(func() {
			cmd.Color = flag.Color{IsSet: true, Value: "true"}
		})

		It("successfully updates the config", func() {
			Expect(executeErr).To(Not(HaveOccurred()))
			Expect(fakeConfig.SetColorEnabledCallCount()).To(Equal(1))
			value := fakeConfig.SetColorEnabledArgsForCall(0)
			Expect(value).To(Equal("true"))
		})
	})

	When("using the locale flag", func() {
		BeforeEach(func() {
			cmd.Locale = flag.Locale{Locale: "en-US"}
		})

		It("successfully updates the config", func() {
			Expect(executeErr).To(Not(HaveOccurred()))
			Expect(fakeConfig.SetLocaleCallCount()).To(Equal(1))
			value := fakeConfig.SetLocaleArgsForCall(0)
			Expect(value).To(Equal("en-US"))
		})
	})

	When("using the trace flag", func() {
		BeforeEach(func() {
			cmd.Trace = "my-trace-file"
		})

		It("successfully updates the config", func() {
			Expect(executeErr).To(Not(HaveOccurred()))
			Expect(fakeConfig.SetTraceCallCount()).To(Equal(1))
			value := fakeConfig.SetTraceArgsForCall(0)
			Expect(value).To(Equal("my-trace-file"))
		})
	})
})
