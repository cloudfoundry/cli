package shared_test

import (
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7/shared"
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

	When("the api endpoint is not set", func() {
		It("returns the NoAPISetError", func() {
			_, _, _, err := GetNewClientsAndConnectToCF(fakeConfig, testUI, "")
			Expect(err).To(MatchError(translatableerror.NoAPISetError{
				BinaryName: binaryName,
			}))
		})
	})

	When("the DialTimeout is set", func() {
		BeforeEach(func() {
			if runtime.GOOS == "windows" {
				Skip("due to timing issues on windows")
			}
			fakeConfig.TargetReturns("https://potato.bananapants11122.co.uk")
			fakeConfig.DialTimeoutReturns(time.Nanosecond)
		})

		It("reads target settings from the config for each client", func() {
			_, _, _, err := GetNewClientsAndConnectToCF(fakeConfig, testUI, "")

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeConfig.TargetCallCount()).To(Equal(2))
			Expect(fakeConfig.DialTimeoutCallCount()).To(Equal(3))
			Expect(fakeConfig.SkipSSLValidationCallCount()).To(Equal(3))
		})
	})

	When("not targeting", func() {
		It("does not target", func() {
			ccClient := NewWrappedCloudControllerClient(fakeConfig, testUI)
			Expect(ccClient).ToNot(BeNil())
			Expect(fakeConfig.SkipSSLValidationCallCount()).To(Equal(0))
		})
	})
})
