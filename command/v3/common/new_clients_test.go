package common_test

import (
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v3/common"
	"code.cloudfoundry.org/cli/utils/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("New Clients", func() {
	var (
		binaryName string
		fakeConfig *commandfakes.FakeConfig
		fakeUI     *ui.UI
	)

	BeforeEach(func() {
		binaryName = "faceman"
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns(binaryName)

		fakeUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
	})

	Context("when the api endpoint is not set", func() {
		It("returns an error", func() {
			_, err := NewClients(fakeConfig, fakeUI)
			Expect(err).To(MatchError(NoAPISetError{
				BinaryName: binaryName,
			}))
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
			_, err := NewClients(fakeConfig, fakeUI)
			Expect(err).To(MatchError("Get https://potato.bananapants11122.co.uk: dial tcp: i/o timeout"))
		})
	})
})
