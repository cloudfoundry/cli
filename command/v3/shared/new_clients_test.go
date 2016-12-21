package shared_test

import (
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("New Clients", func() {
	var (
		binaryName string
		fakeConfig *commandfakes.FakeConfig
		testUI     *ui.UI
	)

	BeforeEach(func() {
		binaryName = "faceman"
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns(binaryName)

		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
	})

	Context("when the api endpoint is not set", func() {
		It("returns the NoAPISetError", func() {
			_, err := NewClients(fakeConfig, testUI)
			Expect(err).To(MatchError(command.NoAPISetError{
				BinaryName: binaryName,
			}))
		})
	})

	Context("when the api does not exist", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("http://4012493825site.com")
		})

		It("returns the ClientTargetError", func() {
			_, err := NewClients(fakeConfig, testUI)
			Expect(err.Error()).To(MatchRegexp("Note that this command requires CF API version 3.0.0+."))
		})
	})

	Context("when the DialTimeout is set", func() {
		BeforeEach(func() {
			if runtime.GOOS == "windows" {
				Skip("due to timing issues on windows")
			}
			fakeConfig.TargetReturns("https://potato.bananapants11122.co.uk")
			fakeConfig.DialTimeoutReturns(time.Nanosecond)
		})

		It("passes the value to the target", func() {
			_, err := NewClients(fakeConfig, testUI)
			if e, ok := err.(ClientTargetError); ok {
				Expect(e.Message).To(MatchRegexp("https://potato.bananapants11122.co.uk: dial tcp.*i/o timeout"))
			} else {
				Fail("Expected err to be type ClientTargetError")
			}
		})
	})
})
