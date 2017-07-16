package shared_test

import (
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v2/shared"
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
		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())

		fakeConfig.BinaryNameReturns(binaryName)
	})

	Context("when the api endpoint is not set", func() {
		It("returns an error", func() {
			_, _, err := NewClients(fakeConfig, testUI, true)
			Expect(err).To(MatchError(translatableerror.NoAPISetError{
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
			_, _, err := NewClients(fakeConfig, testUI, true)
			Expect(err.Error()).To(MatchRegexp("TIP: If you are behind a firewall"))
		})
	})

	Context("when the targeting a CF fails", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("https://potato.bananapants11122.co.uk")
		})

		It("returns an error", func() {
			_, _, err := NewClients(fakeConfig, testUI, true)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when not targetting", func() {
		It("does not target and returns no UAA client", func() {
			ccClient, uaaClient, err := NewClients(fakeConfig, testUI, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ccClient).ToNot(BeNil())
			Expect(uaaClient).To(BeNil())
			Expect(fakeConfig.SkipSSLValidationCallCount()).To(Equal(0))
		})
	})
})
