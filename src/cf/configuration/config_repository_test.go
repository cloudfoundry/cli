package configuration_test

import (
	. "cf/configuration"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testconfig "testhelpers/configuration"
	"testhelpers/maker"
	"time"
)

var _ = Describe("ConfigurationRepository", func() {
	var config Repository
	var repo *testconfig.FakePersistor

	BeforeEach(func() {
		repo = testconfig.NewFakePersistor()
		repo.LoadReturns.Data = NewData()
		config = testconfig.NewRepository()
	})

	It("is safe for concurrent reading and writing", func() {
		swapValLoop := func(config Repository) {
			for {
				val := config.ApiEndpoint()

				switch val {
				case "foo":
					config.SetApiEndpoint("bar")
				case "bar":
					config.SetApiEndpoint("foo")
				default:
					panic(fmt.Sprintf("WAT: %s", val))
				}
			}
		}

		config.SetApiEndpoint("foo")

		go swapValLoop(config)
		go swapValLoop(config)
		go swapValLoop(config)
		go swapValLoop(config)

		time.Sleep(10 * time.Millisecond)
	})

	// TODO - test ClearTokens et al
	It("has acccessor methods for all config fields", func() {
		config.SetApiEndpoint("http://api.the-endpoint")
		assert.Equal(mr.T(), config.ApiEndpoint(), "http://api.the-endpoint")

		config.SetApiVersion("3")
		assert.Equal(mr.T(), config.ApiVersion(), "3")

		config.SetAuthorizationEndpoint("http://auth.the-endpoint")
		assert.Equal(mr.T(), config.AuthorizationEndpoint(), "http://auth.the-endpoint")

		config.SetLoggregatorEndpoint("http://logs.the-endpoint")
		assert.Equal(mr.T(), config.LoggregatorEndpoint(), "http://logs.the-endpoint")

		config.SetAccessToken("the-token")
		assert.Equal(mr.T(), config.AccessToken(), "the-token")

		config.SetRefreshToken("the-token")
		assert.Equal(mr.T(), config.RefreshToken(), "the-token")

		organization := maker.NewOrgFields(maker.Overrides{"name": "the-org"})
		config.SetOrganizationFields(organization)
		assert.Equal(mr.T(), config.OrganizationFields(), organization)

		space := maker.NewSpaceFields(maker.Overrides{"name": "the-space"})
		config.SetSpaceFields(space)
		assert.Equal(mr.T(), config.SpaceFields(), space)
	})

	It("User has a valid Access Token", func() {
		config.SetAccessToken("bearer eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJjNDE4OTllNS1kZTE1LTQ5NGQtYWFiNC04ZmNlYzUxN2UwMDUiLCJzdWIiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJ1c2VyX25hbWUiOiJ1c2VyMUBleGFtcGxlLmNvbSIsImVtYWlsIjoidXNlcjFAZXhhbXBsZS5jb20iLCJpYXQiOjEzNzcwMjgzNTYsImV4cCI6MTM3NzAzNTU1NiwiaXNzIjoiaHR0cHM6Ly91YWEuYXJib3JnbGVuLmNmLWFwcC5jb20vb2F1dGgvdG9rZW4iLCJhdWQiOlsib3BlbmlkIiwiY2xvdWRfY29udHJvbGxlciIsInBhc3N3b3JkIl19.kjFJHi0Qir9kfqi2eyhHy6kdewhicAFu8hrPR1a5AxFvxGB45slKEjuP0_72cM_vEYICgZn3PcUUkHU9wghJO9wjZ6kiIKK1h5f2K9g-Iprv9BbTOWUODu1HoLIvg2TtGsINxcRYy_8LW1RtvQc1b4dBPoopaEH4no-BIzp0E5E")
		assert.Equal(mr.T(), config.UserGuid(), "772dda3f-669f-4276-b2bd-90486abe1f6f")
		assert.Equal(mr.T(), config.UserEmail(), "user1@example.com")
	})

	It("User has an invalid Access Token", func() {
		config.SetAccessToken("bearer")
		assert.Empty(mr.T(), config.UserGuid())
		assert.Empty(mr.T(), config.UserEmail())

		config.SetAccessToken("bearer eyJhbGciOiJSUzI1NiJ9")
		assert.Empty(mr.T(), config.UserGuid())
		assert.Empty(mr.T(), config.UserEmail())
	})
})
