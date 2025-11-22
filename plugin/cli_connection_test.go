package plugin_test

import (
	"github.com/cloudfoundry/cli/plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CliConnection", func() {
	Describe("NewCliConnection", func() {
		It("creates a new CLI connection with the specified port", func() {
			conn := plugin.NewCliConnection("12345")
			Expect(conn).ToNot(BeNil())
		})

		It("creates connection with empty port", func() {
			conn := plugin.NewCliConnection("")
			Expect(conn).ToNot(BeNil())
		})

		It("creates connection with different ports", func() {
			conn1 := plugin.NewCliConnection("8080")
			conn2 := plugin.NewCliConnection("9090")
			Expect(conn1).ToNot(BeNil())
			Expect(conn2).ToNot(BeNil())
		})
	})

	// Note: Full RPC testing requires a test RPC server setup
	// See plugin_shim_test.go for examples of integration tests with RPC
	// These tests verify the structure and basic functionality
	// Integration tests with actual RPC communication are in plugin_shim_test.go

	Describe("RPC Methods Structure", func() {
		var conn plugin.CliConnectionInterface

		BeforeEach(func() {
			conn = plugin.NewCliConnection("12345")
		})

		Context("when RPC server is not available", func() {
			It("CliCommand returns error when cannot connect", func() {
				_, err := conn.CliCommand("apps")
				Expect(err).To(HaveOccurred())
			})

			It("CliCommandWithoutTerminalOutput returns error when cannot connect", func() {
				_, err := conn.CliCommandWithoutTerminalOutput("apps")
				Expect(err).To(HaveOccurred())
			})

			It("GetCurrentOrg returns error when cannot connect", func() {
				_, err := conn.GetCurrentOrg()
				Expect(err).To(HaveOccurred())
			})

			It("GetCurrentSpace returns error when cannot connect", func() {
				_, err := conn.GetCurrentSpace()
				Expect(err).To(HaveOccurred())
			})

			It("Username returns error when cannot connect", func() {
				_, err := conn.Username()
				Expect(err).To(HaveOccurred())
			})

			It("UserGuid returns error when cannot connect", func() {
				_, err := conn.UserGuid()
				Expect(err).To(HaveOccurred())
			})

			It("UserEmail returns error when cannot connect", func() {
				_, err := conn.UserEmail()
				Expect(err).To(HaveOccurred())
			})

			It("IsSSLDisabled returns error when cannot connect", func() {
				_, err := conn.IsSSLDisabled()
				Expect(err).To(HaveOccurred())
			})

			It("IsLoggedIn returns error when cannot connect", func() {
				_, err := conn.IsLoggedIn()
				Expect(err).To(HaveOccurred())
			})

			It("HasOrganization returns error when cannot connect", func() {
				_, err := conn.HasOrganization()
				Expect(err).To(HaveOccurred())
			})

			It("HasSpace returns error when cannot connect", func() {
				_, err := conn.HasSpace()
				Expect(err).To(HaveOccurred())
			})

			It("ApiEndpoint returns error when cannot connect", func() {
				_, err := conn.ApiEndpoint()
				Expect(err).To(HaveOccurred())
			})

			It("HasAPIEndpoint returns error when cannot connect", func() {
				_, err := conn.HasAPIEndpoint()
				Expect(err).To(HaveOccurred())
			})

			It("ApiVersion returns error when cannot connect", func() {
				_, err := conn.ApiVersion()
				Expect(err).To(HaveOccurred())
			})

			It("LoggregatorEndpoint returns error when cannot connect", func() {
				_, err := conn.LoggregatorEndpoint()
				Expect(err).To(HaveOccurred())
			})

			It("DopplerEndpoint returns error when cannot connect", func() {
				_, err := conn.DopplerEndpoint()
				Expect(err).To(HaveOccurred())
			})

			It("AccessToken returns error when cannot connect", func() {
				_, err := conn.AccessToken()
				Expect(err).To(HaveOccurred())
			})

			It("GetApp returns error when cannot connect", func() {
				_, err := conn.GetApp("my-app")
				Expect(err).To(HaveOccurred())
			})

			It("GetApps returns error when cannot connect", func() {
				_, err := conn.GetApps()
				Expect(err).To(HaveOccurred())
			})

			It("GetOrgs returns error when cannot connect", func() {
				_, err := conn.GetOrgs()
				Expect(err).To(HaveOccurred())
			})

			It("GetSpaces returns error when cannot connect", func() {
				_, err := conn.GetSpaces()
				Expect(err).To(HaveOccurred())
			})

			It("GetServices returns error when cannot connect", func() {
				_, err := conn.GetServices()
				Expect(err).To(HaveOccurred())
			})

			It("GetOrgUsers returns error when cannot connect", func() {
				_, err := conn.GetOrgUsers("my-org")
				Expect(err).To(HaveOccurred())
			})

			It("GetSpaceUsers returns error when cannot connect", func() {
				_, err := conn.GetSpaceUsers("my-org", "my-space")
				Expect(err).To(HaveOccurred())
			})

			It("GetOrg returns error when cannot connect", func() {
				_, err := conn.GetOrg("my-org")
				Expect(err).To(HaveOccurred())
			})

			It("GetSpace returns error when cannot connect", func() {
				_, err := conn.GetSpace("my-space")
				Expect(err).To(HaveOccurred())
			})

			It("GetService returns error when cannot connect", func() {
				_, err := conn.GetService("my-service")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Method Signatures", func() {
		It("CliCommand accepts variadic string arguments", func() {
			conn := plugin.NewCliConnection("12345")
			_, err := conn.CliCommand("apps")
			Expect(err).To(HaveOccurred()) // Will fail to connect, but verifies signature

			_, err = conn.CliCommand("push", "my-app")
			Expect(err).To(HaveOccurred())

			_, err = conn.CliCommand("set-env", "my-app", "KEY", "value")
			Expect(err).To(HaveOccurred())
		})

		It("CliCommandWithoutTerminalOutput accepts variadic string arguments", func() {
			conn := plugin.NewCliConnection("12345")
			_, err := conn.CliCommandWithoutTerminalOutput("apps")
			Expect(err).To(HaveOccurred())

			_, err = conn.CliCommandWithoutTerminalOutput("app", "my-app")
			Expect(err).To(HaveOccurred())
		})

		It("GetOrgUsers accepts orgName and variadic args", func() {
			conn := plugin.NewCliConnection("12345")
			_, err := conn.GetOrgUsers("my-org")
			Expect(err).To(HaveOccurred())

			_, err = conn.GetOrgUsers("my-org", "arg1", "arg2")
			Expect(err).To(HaveOccurred())
		})
	})
})
