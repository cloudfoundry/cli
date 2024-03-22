package configv3_test

import (
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DefaultUserConfig", func() {
	var config configv3.DefaultUserConfig

	Describe("CurrentUser", func() {
		When("using client credentials and the user token is set", func() {
			It("returns the user", func() {
				config = configv3.DefaultUserConfig{
					ConfigFile: &configv3.JSONConfig{
						AccessToken: AccessTokenForClientUsers,
					},
				}
				user, err := config.CurrentUser()
				Expect(err).ToNot(HaveOccurred())
				Expect(user).To(Equal(configv3.User{
					Name:     "potato-face",
					GUID:     "potato-face",
					IsClient: true,
				}))
			})
		})

		When("using user/password and the user token is set", func() {
			It("returns the user", func() {
				config = configv3.DefaultUserConfig{
					ConfigFile: &configv3.JSONConfig{
						AccessToken: AccessTokenForHumanUsers,
					},
				}

				user, err := config.CurrentUser()
				Expect(err).ToNot(HaveOccurred())
				Expect(user).To(Equal(configv3.User{
					Name:     "admin",
					GUID:     "9519be3e-44d9-40d0-ab9a-f4ace11df159",
					Origin:   "uaa",
					IsClient: false,
				}))
			})
		})

		When("the user token is blank", func() {
			It("returns the user", func() {
				config = configv3.DefaultUserConfig{ConfigFile: &configv3.JSONConfig{}}

				user, err := config.CurrentUser()
				Expect(err).ToNot(HaveOccurred())
				Expect(user).To(Equal(configv3.User{}))
			})
		})
	})

	Describe("CurrentUserName", func() {
		When("using client credentials and the user token is set", func() {
			It("returns the username", func() {
				config = configv3.DefaultUserConfig{
					ConfigFile: &configv3.JSONConfig{
						AccessToken: AccessTokenForClientUsers,
					},
				}

				username, err := config.CurrentUserName()
				Expect(err).ToNot(HaveOccurred())
				Expect(username).To(Equal("potato-face"))
			})
		})

		When("using user/password and the user token is set", func() {
			It("returns the username", func() {
				config = configv3.DefaultUserConfig{
					ConfigFile: &configv3.JSONConfig{
						AccessToken: AccessTokenForHumanUsers,
					},
				}

				username, err := config.CurrentUserName()
				Expect(err).ToNot(HaveOccurred())
				Expect(username).To(Equal("admin"))
			})
		})

		When("the user token is blank", func() {
			It("returns an empty string", func() {
				config = configv3.DefaultUserConfig{ConfigFile: &configv3.JSONConfig{}}
				username, err := config.CurrentUserName()
				Expect(err).ToNot(HaveOccurred())
				Expect(username).To(BeEmpty())
			})
		})
	})
})
