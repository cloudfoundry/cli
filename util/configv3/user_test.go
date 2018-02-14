package configv3_test

import (
	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var homeDir string

	BeforeEach(func() {
		homeDir = setup()
	})

	AfterEach(func() {
		teardown(homeDir)
	})

	PDescribe("CurrentUser", func() {
		Context("when using client credentials and the user token is set", func() {
			It("returns the user", func() {
				config := Config{
					ConfigFile: JSONConfig{
						AccessToken: "bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleS0xIiwidHlwIjoiSldUIn0.eyJqdGkiOiI5MTExMzczOTRjYTg0NzQzOGUxZjQyOWY4OTQ2ZGZmMyIsInN1YiI6InBvdGF0by1mYWNlIiwiYXV0aG9yaXRpZXMiOlsicm91dGluZy5yb3V0ZXJfZ3JvdXBzLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJwYXNzd29yZC53cml0ZSIsImNsb3VkX2NvbnRyb2xsZXIud3JpdGUiLCJvcGVuaWQiLCJuZXR3b3JrLmFkbWluIiwicm91dGluZy5yb3V0ZXJfZ3JvdXBzLndyaXRlIiwiZG9wcGxlci5maXJlaG9zZSIsInNjaW0ud3JpdGUiLCJzY2ltLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLmFkbWluIiwidWFhLnVzZXIiXSwic2NvcGUiOlsicm91dGluZy5yb3V0ZXJfZ3JvdXBzLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJwYXNzd29yZC53cml0ZSIsImNsb3VkX2NvbnRyb2xsZXIud3JpdGUiLCJvcGVuaWQiLCJuZXR3b3JrLmFkbWluIiwicm91dGluZy5yb3V0ZXJfZ3JvdXBzLndyaXRlIiwiZG9wcGxlci5maXJlaG9zZSIsInNjaW0ud3JpdGUiLCJzY2ltLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLmFkbWluIiwidWFhLnVzZXIiXSwiY2xpZW50X2lkIjoicG90YXRvLWZhY2UiLCJjaWQiOiJwb3RhdG8tZmFjZSIsImF6cCI6InBvdGF0by1mYWNlIiwiZ3JhbnRfdHlwZSI6ImNsaWVudF9jcmVkZW50aWFscyIsInJldl9zaWciOiJkMjU1NjdjYiIsImlhdCI6MTUxNjg0MTY4MCwiZXhwIjoxNTE2ODQxNzQwLCJpc3MiOiJodHRwczovL3VhYS5ib3NoLWxpdGUuY29tL29hdXRoL3Rva2VuIiwiemlkIjoidWFhIiwiYXVkIjpbImNsb3VkX2NvbnRyb2xsZXIiLCJzY2ltIiwicG90YXRvLWZhY2UiLCJwYXNzd29yZCIsInVhYSIsIm9wZW5pZCIsImRvcHBsZXIiLCJyb3V0aW5nLnJvdXRlcl9ncm91cHMiLCJuZXR3b3JrIl19.I2flQSfAhWiCdhyd0414dZ0qmv0W-dPTGvj0pIXnaFPNae7gXSz79MPipTelSxCvdtigX8SoW8O7dWU5zt0O7VRkQX_YYElTHnQeWBfljoFvHhYPRMUv24I3lO6beeujKlYbUxVP5BXoyEdyfiDwzJjoX9lzxriBKdY_BO81oRUjItl7oI1VFhj1A_UwUcDwK2t-c7zDxSmh4P48r77QdqDoAjuweZPUU4PdzRlp99XYdmke52KeG7Xums6hrEJJBSDLbczd_308FAitaQKHgAQH1swLQXqcuD29-eoB4_nTSBwok2H5hoHicHWohSBkMOyFqD4HPb8ta8d_FOi8HA",
					},
				}

				user, err := config.CurrentUser()
				Expect(err).ToNot(HaveOccurred())
				Expect(user).To(Equal(User{
					Name: "potato-face",
				}))
			})
		})

		Context("when using user/password and the user token is set", func() {
			It("returns the user", func() {
				config := Config{
					ConfigFile: JSONConfig{
						AccessToken: "bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImxlZ2FjeS10b2tlbi1rZXkiLCJ0eXAiOiJKV1QifQ.eyJqdGkiOiI3YzZkMDA2MjA2OTI0NmViYWI0ZjBmZjY3NGQ3Zjk4OSIsInN1YiI6Ijk1MTliZTNlLTQ0ZDktNDBkMC1hYjlhLWY0YWNlMTFkZjE1OSIsInNjb3BlIjpbIm9wZW5pZCIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy53cml0ZSIsInNjaW0ucmVhZCIsImNsb3VkX2NvbnRyb2xsZXIuYWRtaW4iLCJ1YWEudXNlciIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy5yZWFkIiwiY2xvdWRfY29udHJvbGxlci5yZWFkIiwicGFzc3dvcmQud3JpdGUiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwiZG9wcGxlci5maXJlaG9zZSIsInNjaW0ud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImF6cCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6Ijk1MTliZTNlLTQ0ZDktNDBkMC1hYjlhLWY0YWNlMTFkZjE1OSIsIm9yaWdpbiI6InVhYSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsImF1dGhfdGltZSI6MTQ3MzI4NDU3NywicmV2X3NpZyI6IjZiMjdkYTZjIiwiaWF0IjoxNDczMjg0NTc3LCJleHAiOjE0NzMyODUxNzcsImlzcyI6Imh0dHBzOi8vdWFhLmJvc2gtbGl0ZS5jb20vb2F1dGgvdG9rZW4iLCJ6aWQiOiJ1YWEiLCJhdWQiOlsiY2YiLCJvcGVuaWQiLCJyb3V0aW5nLnJvdXRlcl9ncm91cHMiLCJzY2ltIiwiY2xvdWRfY29udHJvbGxlciIsInVhYSIsInBhc3N3b3JkIiwiZG9wcGxlciJdfQ.OcH_w9yIKJkEcTZMThIs-qJAHk3G0JwNjG-aomVH9hKye4ciFO6IMQMLKmCBrrAQVc7ST1SZZwq7gv12Dq__6Jp-hai0a2_ADJK-Vc9YXyNZKgYTWIeVNGM1JGdHgFSrBR2Lz7IIrH9HqeN8plrKV5HzU8uI9LL4lyOCjbXJ9cM",
					},
				}

				user, err := config.CurrentUser()
				Expect(err).ToNot(HaveOccurred())
				Expect(user).To(Equal(User{
					Name: "admin",
				}))
			})
		})

		Context("when the user token is blank", func() {
			It("returns the user", func() {
				var config Config
				user, err := config.CurrentUser()
				Expect(err).ToNot(HaveOccurred())
				Expect(user).To(Equal(User{}))
			})
		})
	})
})
