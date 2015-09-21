package credential_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/credential"
	"github.com/cloudfoundry/cli/plugin/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Credential", func() {
	var (
		fakeCliConnection *fakes.FakeCliConnection
		credFactory       credential.CredentialFactory
	)

	BeforeEach(func() {
		fakeCliConnection = &fakes.FakeCliConnection{}
		credFactory = credential.NewCredentialFactory(fakeCliConnection)
	})

	Describe("Get", func() {
		var expectedResponse []string

		Context("when retrieving /v2/info is successful", func() {
			BeforeEach(func() {
				expectedResponse = []string{
					"Getting OAuth token\n",
					"OK\n",
					"bearer lives_in_a_man_cave",
				}

				fakeCliConnection.CliCommandWithoutTerminalOutputReturns(expectedResponse, nil)
			})

			It("returns a populated Info model", func() {
				cred, err := credFactory.Get()
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeCliConnection.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))
				Expect(fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(0)).To(ConsistOf("oauth-token"))

				Expect(cred.Token).To(Equal("bearer lives_in_a_man_cave"))
			})
		})

		Context("when getting the oauth-token fails", func() {
			BeforeEach(func() {
				fakeCliConnection.CliCommandWithoutTerminalOutputReturns(nil, errors.New("woops"))
			})

			It("fails with an error", func() {
				_, err := credFactory.Get()

				Expect(fakeCliConnection.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))
				Expect(err).To(MatchError("Failed to acquire oauth token"))
			})
		})
	})
})
