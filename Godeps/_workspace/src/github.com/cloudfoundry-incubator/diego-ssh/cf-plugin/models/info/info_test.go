package info_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/info"
	"github.com/cloudfoundry/cli/plugin/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Info", func() {
	var (
		fakeCliConnection *fakes.FakeCliConnection
		infoFactory       info.InfoFactory
	)

	BeforeEach(func() {
		fakeCliConnection = &fakes.FakeCliConnection{}
		infoFactory = info.NewInfoFactory(fakeCliConnection)
	})

	Describe("Get", func() {
		var expectedJson string

		JustBeforeEach(func() {
			fakeCliConnection.CliCommandWithoutTerminalOutputReturns([]string{expectedJson}, nil)
		})

		Context("when retrieving /v2/info is successful", func() {
			BeforeEach(func() {
				expectedJson = `{
					"app_ssh_endpoint": "ssh.example.com:1234",
					"app_ssh_host_key_fingerprint": "00:11:22:33:44:55:66:77:88"
				}`

				fakeCliConnection.CliCommandWithoutTerminalOutputReturns([]string{expectedJson}, nil)
			})

			It("returns a populated Info model", func() {
				model, err := infoFactory.Get()
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeCliConnection.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))
				Expect(fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(0)).To(ConsistOf("curl", "/v2/info"))

				Expect(model.SSHEndpoint).To(Equal("ssh.example.com:1234"))
				Expect(model.SSHEndpointFingerprint).To(Equal("00:11:22:33:44:55:66:77:88"))
			})
		})

		Context("when retrieving /v2/info fails", func() {
			JustBeforeEach(func() {
				fakeCliConnection.CliCommandWithoutTerminalOutputReturns(nil, errors.New("woops"))
			})

			It("fails with an error", func() {
				_, err := infoFactory.Get()
				Expect(err).To(MatchError("Failed to acquire SSH endpoint info"))
			})
		})

		Context("when the json response fails to unmarshal", func() {
			BeforeEach(func() {
				expectedJson = `soo, this is bad #json`
			})

			It("fails with an error", func() {
				_, err := infoFactory.Get()
				Expect(err).To(MatchError("Failed to acquire SSH endpoint info"))
			})
		})
	})
})
