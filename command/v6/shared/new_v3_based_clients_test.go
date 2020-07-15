package shared_test

import (
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("New V3 Based Clients", func() {
	var (
		binaryName string
		fakeConfig *commandfakes.FakeConfig
		testUI     *ui.UI
	)

	BeforeEach(func() {
		binaryName = "faceman"
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.UAAEndpointReturns("uaa.com")
		fakeConfig.AuthorizationEndpointReturns("auth.com")

		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
	})

	When("the api endpoint is not set", func() {
		It("returns the NoAPISetError", func() {
			_, _, err := NewV3BasedClients(fakeConfig, testUI, true)
			Expect(err).To(MatchError(translatableerror.NoAPISetError{
				BinaryName: binaryName,
			}))
		})
	})

	When("not targetting", func() {
		It("does not target and returns no UAA client", func() {
			ccClient, uaaClient, err := NewV3BasedClients(fakeConfig, testUI, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ccClient).ToNot(BeNil())
			Expect(uaaClient).To(BeNil())
			Expect(fakeConfig.SkipSSLValidationCallCount()).To(Equal(0))
		})
	})
})
