package common_test

import (
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/commands/commandsfakes"
	. "code.cloudfoundry.org/cli/commands/v3/common"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("New Clients", func() {
	var (
		binaryName string
		fakeConfig *commandsfakes.FakeConfig
	)

	BeforeEach(func() {
		binaryName = "faceman"
		fakeConfig = new(commandsfakes.FakeConfig)
		fakeConfig.BinaryNameReturns(binaryName)
	})

	Context("when the api endpoint is not set", func() {
		It("returns an error", func() {
			_, err := NewClients(fakeConfig)
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
			_, err := NewClients(fakeConfig)
			Expect(err).To(MatchError("Get https://potato.bananapants11122.co.uk: dial tcp: i/o timeout"))
		})
	})
})
