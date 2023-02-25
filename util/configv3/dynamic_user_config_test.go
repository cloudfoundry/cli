package configv3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/configv3/configv3fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DynamicUserConfig", func() {
	var (
		fakeDefaultUserConfig    *configv3fakes.FakeUserConfig
		fakeKubernetesUserConfig *configv3fakes.FakeUserConfig
		jsonConfig               *configv3.JSONConfig
		dynamicUserConfig        configv3.DynamicUserConfig
		err                      error
	)

	BeforeEach(func() {
		fakeDefaultUserConfig = new(configv3fakes.FakeUserConfig)
		fakeDefaultUserConfig.CurrentUserReturns(configv3.User{Name: "default-user"}, nil)
		fakeDefaultUserConfig.CurrentUserNameReturns("default-user", nil)

		fakeKubernetesUserConfig = new(configv3fakes.FakeUserConfig)
		fakeKubernetesUserConfig.CurrentUserReturns(configv3.User{Name: "kubernetes-user"}, nil)
		fakeKubernetesUserConfig.CurrentUserNameReturns("kubernetes-user", nil)

		jsonConfig = &configv3.JSONConfig{}
		dynamicUserConfig = configv3.DynamicUserConfig{
			ConfigFile:           jsonConfig,
			DefaultUserConfig:    fakeDefaultUserConfig,
			KubernetesUserConfig: fakeKubernetesUserConfig,
		}
	})

	Describe("CurrentUser", func() {
		var currentUser configv3.User

		JustBeforeEach(func() {
			currentUser, err = dynamicUserConfig.CurrentUser()
		})

		When("using a default config", func() {
			BeforeEach(func() {
				jsonConfig.CFOnK8s.Enabled = false
			})

			It("delegates to the default UserConfig", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(currentUser.Name).To(Equal("default-user"))
			})

			When("the default UserConfig fails", func() {
				BeforeEach(func() {
					fakeDefaultUserConfig.CurrentUserReturns(configv3.User{}, errors.New("current-user-err"))
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("current-user-err"))
				})
			})
		})

		When("using a Kubernetes config", func() {
			BeforeEach(func() {
				jsonConfig.CFOnK8s.Enabled = true
			})

			It("delegates to the Kubernetes UserConfig", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(currentUser.Name).To(Equal("kubernetes-user"))
			})

			When("the Kubernetes UserConfig fails", func() {
				BeforeEach(func() {
					fakeKubernetesUserConfig.CurrentUserReturns(configv3.User{}, errors.New("current-user-err"))
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("current-user-err"))
				})
			})
		})
	})

	Describe("CurrentUserName", func() {
		var currentUserName string

		JustBeforeEach(func() {
			currentUserName, err = dynamicUserConfig.CurrentUserName()
		})

		When("using a default config", func() {
			BeforeEach(func() {
				jsonConfig.CFOnK8s.Enabled = false
			})

			It("delegates to the default UserConfig", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(currentUserName).To(Equal("default-user"))
			})

			When("the default UserConfig fails", func() {
				BeforeEach(func() {
					fakeDefaultUserConfig.CurrentUserNameReturns("", errors.New("current-username-err"))
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("current-username-err"))
				})
			})
		})

		When("using a Kubernetes config", func() {
			BeforeEach(func() {
				jsonConfig.CFOnK8s.Enabled = true
			})

			It("delegates to the Kubernetes UserConfig", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(currentUserName).To(Equal("kubernetes-user"))
			})

			When("the Kubernetes UserConfig fails", func() {
				BeforeEach(func() {
					fakeKubernetesUserConfig.CurrentUserNameReturns("", errors.New("current-username-err"))
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("current-username-err"))
				})
			})
		})
	})
})
