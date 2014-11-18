// +build darwin freebsd linux netbsd openbsd

package i18n_test

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/i18n"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	go_i18n "github.com/nicksnyder/go-i18n/i18n"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("i18n.Init() function", func() {
	var (
		oldResourcesPath string
		configRepo       core_config.ReadWriter

		T go_i18n.TranslateFunc
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		oldResourcesPath = i18n.GetResourcesPath()
		i18n.Resources_path = filepath.Join("cf", "i18n", "test_fixtures")
	})

	JustBeforeEach(func() {
		T = i18n.Init(configRepo)
	})

	Describe("When a user has a locale configuration set", func() {
		It("panics when the translation files cannot be loaded", func() {
			i18n.Resources_path = filepath.Join("should", "not", "be_valid")
			configRepo.SetLocale("en_us")

			init := func() { i18n.Init(configRepo) }
			Ω(init).Should(Panic(), "loading translations from an invalid path should panic")
		})

		It("Panics if the locale is not valid", func() {
			configRepo.SetLocale("abc_def")

			init := func() { i18n.Init(configRepo) }
			Ω(init).Should(Panic(), "loading translations from an invalid path should panic")
		})

		Context("when the locale is set to french", func() {
			BeforeEach(func() {
				configRepo.SetLocale("fr_FR")
			})

			It("translates into french correctly", func() {
				translation := T("No buildpacks found")
				Ω(translation).Should(Equal("Pas buildpacks trouvés"))
			})
		})

		Context("creates a valid T function", func() {
			BeforeEach(func() {
				configRepo.SetLocale("en_US")
			})

			It("returns a usable T function for simple strings", func() {
				Ω(T).ShouldNot(BeNil())

				translation := T("Hello world!")
				Ω("Hello world!").Should(Equal(translation))
			})

			It("returns a usable T function for complex strings (interpolated)", func() {
				Ω(T).ShouldNot(BeNil())

				translation := T("Deleting domain {{.DomainName}} as {{.Username}}...", map[string]interface{}{"DomainName": "foo.com", "Username": "Anand"})
				Ω("Deleting domain foo.com as Anand...").Should(Equal(translation))
			})
		})
	})

	Describe("When the user does not have a locale configuration set", func() {
		AfterEach(func() {
			i18n.Resources_path = oldResourcesPath
			os.Setenv("LC_ALL", "")
			os.Setenv("LANG", "en_US.UTF-8")
		})

		It("panics when the translation files cannot be loaded", func() {
			os.Setenv("LANG", "en")
			i18n.Resources_path = filepath.Join("should", "not", "be_valid")

			init := func() { i18n.Init(configRepo) }
			Ω(init).Should(Panic(), "loading translations from an invalid path should panic")
		})

		Context("loads correct locale", func() {
			It("defaults to en_US when LC_ALL and LANG not set", func() {
				os.Setenv("LC_ALL", "")
				os.Setenv("LANG", "")

				translation := T("Hello world!")
				Ω("Hello world!").Should(Equal(translation))
			})

			Context("when there is no territory set", func() {
				BeforeEach(func() {
					os.Setenv("LANG", "en")
				})

				It("still loads the english translation", func() {
					translation := T("Hello world!")
					Ω("Hello world!").Should(Equal(translation))
				})
			})

			Context("when the desired language is not supported", func() {
				BeforeEach(func() {
					os.Setenv("LC_ALL", "zz_FF.UTF-8")
				})

				It("defaults to en_US when langauge is not supported", func() {
					translation := T("Hello world!")
					Ω("Hello world!").Should(Equal(translation))

					translation = T("No buildpacks found")
					Ω("No buildpacks found").Should(Equal(translation))
				})

				Context("because we don't have the territory", func() {
					BeforeEach(func() {
						os.Setenv("LC_ALL", "fr_CA.UTF-8")
					})

					It("defaults to same language in supported territory", func() {
						translation := T("No buildpacks found")
						Ω("Pas buildpacks trouvés").Should(Equal(translation))
					})
				})
			})

			Context("translates correctly", func() {
				BeforeEach(func() {
					os.Setenv("LC_ALL", "fr_FR.UTF-8")
				})

				It("T function should return translation if string key exists", func() {
					translation := T("No buildpacks found")
					Ω("Pas buildpacks trouvés").Should(Equal(translation))
				})
			})

			Context("matches zh_CN to simplified Chinese", func() {
				BeforeEach(func() {
					os.Setenv("LC_ALL", "zh_CN.UTF-8")
				})

				It("matches to zh_Hans", func() {
					translation := T("No buildpacks found")
					Ω("buildpack未找到").Should(Equal(translation))
				})
			})

			Context("matches zh_TW locale to traditional Chinese", func() {
				BeforeEach(func() {
					os.Setenv("LC_ALL", "zh_TW.UTF-8")
				})

				It("matches to zh_Hant", func() {
					translation := T("No buildpacks found")
					Ω("(Hant)No buildpacks found").Should(Equal(translation))
				})
			})

			Context("matches zh_HK locale to traditional Chinese", func() {
				BeforeEach(func() {
					os.Setenv("LC_ALL", "zh_HK.UTF-8")
				})

				It("matches to zh_Hant", func() {
					translation := T("No buildpacks found")
					Ω("(Hant)No buildpacks found").Should(Equal(translation))
				})
			})
		})

		Context("creates a valid T function", func() {
			BeforeEach(func() {
				os.Setenv("LC_ALL", "en_US.UTF-8")
			})

			It("returns a usable T function for simple strings", func() {
				Ω(T).ShouldNot(BeNil())

				translation := T("Hello world!")
				Ω("Hello world!").Should(Equal(translation))
			})

			It("returns a usable T function for complex strings (interpolated)", func() {
				Ω(T).ShouldNot(BeNil())

				translation := T("Deleting domain {{.DomainName}} as {{.Username}}...", map[string]interface{}{"DomainName": "foo.com", "Username": "Anand"})
				Ω("Deleting domain foo.com as Anand...").Should(Equal(translation))
			})
		})
	})

	Describe("when the config is set to a non-english language and the LANG environamnt variable is en_US", func() {
		BeforeEach(func() {
			configRepo.SetLocale("fr_FR")
			os.Setenv("LANG", "en_US")
		})

		AfterEach(func() {
			i18n.Resources_path = oldResourcesPath
			os.Setenv("LANG", "en_US.UTF-8")
		})

		It("ignores the english LANG enviornmant variable", func() {
			translation := T("No buildpacks found")
			Ω(translation).Should(Equal("Pas buildpacks trouvés"))
		})
	})
})
