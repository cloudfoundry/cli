// +build darwin freebsd linux netbsd openbsd

package i18n_test

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/i18n/detection"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("i18n.Init() function", func() {
	var (
		oldResourcesPath string
		configRepo       core_config.ReadWriter
		detector         detection.Detector
	)

	BeforeEach(func() {
		oldResourcesPath = i18n.GetResourcesPath()
		i18n.Resources_path = filepath.Join("cf", "i18n", "test_fixtures")
		configRepo = testconfig.NewRepositoryWithDefaults()
		detector = &detection.JibberJabberDetector{}
	})

	AfterEach(func() {
		i18n.Resources_path = oldResourcesPath
	})

	Describe("Init", func() {
		Context("when the config contains a locale", func() {
			BeforeEach(func() {
				configRepo.SetLocale("fr_FR")
			})

			Context("when the translations can be loaded", func() {
				It("returns a translate func for that locale", func() {
					translateFunc := i18n.Init(configRepo, detector)
					Expect(translateFunc("No buildpacks found")).To(Equal("Pas buildpacks trouvés"))
				})
			})

			Context("when the translations cannot be loaded", func() {
				// no file exists for that locale
				// file when loaded returned no bytes
				// root tempdir could not be found
				// os.Stat failed for a reason other than the dir does not exist
				// creating the tempdir failed
				// creating the tempfile failed
				// writing the tempfile failed
				// go_i18n could not load the tempfile

				BeforeEach(func() {
					configRepo.SetLocale("invalid")
				})

				It("panics", func() {
					Expect(func() { i18n.Init(configRepo, detector) }).To(Panic())
				})
			})
		})

		Context("when the config does not contain a locale", func() {
			BeforeEach(func() {
				configRepo.SetLocale("")
			})

			Context("when LANG and LC_ALL are not set", func() {
				var origLANG, origLCALL string

				BeforeEach(func() {
					origLANG = os.Getenv("LANG")
					origLCALL = os.Getenv("LC_ALL")
					os.Setenv("LANG", "")
					os.Setenv("LC_ALL", "")
				})

				AfterEach(func() {
					os.Setenv("LANG", origLANG)
					os.Setenv("LC_ALL", origLCALL)
				})

				It("returns a translate func for the default locale", func() {
					translateFunc := i18n.Init(configRepo, detector)
					Expect(translateFunc("No buildpacks found")).To(Equal("No buildpacks found"))
				})
			})

			Context("when LANG is set to zh_TW", func() {
				var origLANG, origLCALL string

				BeforeEach(func() {
					origLANG = os.Getenv("LANG")
					origLCALL = os.Getenv("LC_ALL")
					os.Setenv("LANG", "zh_TW")
					os.Setenv("LC_ALL", "")
				})

				AfterEach(func() {
					os.Setenv("LANG", origLANG)
					os.Setenv("LC_ALL", origLCALL)
				})

				It("returns a translate func for zh_Hant", func() {
					translateFunc := i18n.Init(configRepo, detector)
					Expect(translateFunc("No buildpacks found")).To(Equal("(Hant)No buildpacks found"))
				})
			})

			Context("when LANG is set to zh_HK", func() {
				var origLANG, origLCALL string

				BeforeEach(func() {
					origLANG = os.Getenv("LANG")
					origLCALL = os.Getenv("LC_ALL")
					os.Setenv("LANG", "zh_HK")
					os.Setenv("LC_ALL", "")
				})

				AfterEach(func() {
					os.Setenv("LANG", origLANG)
					os.Setenv("LC_ALL", origLCALL)
				})

				It("returns a translate func for zh_Hant", func() {
					translateFunc := i18n.Init(configRepo, detector)
					Expect(translateFunc("No buildpacks found")).To(Equal("(Hant)No buildpacks found"))
				})
			})

			Context("when LANG is set to something other than zh_TW or zh_HK", func() {
				Context("when the language can be loaded via locale", func() {
					var origLANG, origLCALL string

					BeforeEach(func() {
						origLANG = os.Getenv("LANG")
						origLCALL = os.Getenv("LC_ALL")
						os.Setenv("LANG", "fr_FR")
						os.Setenv("LC_ALL", "")
					})

					AfterEach(func() {
						os.Setenv("LANG", origLANG)
						os.Setenv("LC_ALL", origLCALL)
					})

					It("returns a translate func for that locale", func() {
						translateFunc := i18n.Init(configRepo, detector)
						Expect(translateFunc("No buildpacks found")).To(Equal("Pas buildpacks trouvés"))
					})
				})

				Context("when the translations can be loaded via language", func() {
					var origLANG, origLCALL string

					BeforeEach(func() {
						origLANG = os.Getenv("LANG")
						origLCALL = os.Getenv("LC_ALL")
						os.Setenv("LANG", "fr_ZZ")
						os.Setenv("LC_ALL", "")
					})

					AfterEach(func() {
						os.Setenv("LANG", origLANG)
						os.Setenv("LC_ALL", origLCALL)
					})

					Context("when the translations can be loaded", func() {
						It("returns a translate func for that locale", func() {
							translateFunc := i18n.Init(configRepo, detector)
							Expect(translateFunc("No buildpacks found")).To(Equal("Pas buildpacks trouvés"))
						})
					})

					Context("when the translations cannot be loaded", func() {
						It("returns a translate func for the default locale", func() {})
					})
				})

				Context("when the translations cannot be found", func() {
					var origLANG string

					BeforeEach(func() {
						origLANG = os.Getenv("LANG")
						os.Setenv("LANG", "zz_ZZ")
					})

					AfterEach(func() {
						os.Setenv("LANG", origLANG)
					})

					It("returns a translate func for the default locale", func() {
						translateFunc := i18n.Init(configRepo, detector)
						Expect(translateFunc("No buildpacks found")).To(Equal("No buildpacks found"))
					})
				})
			})

			Context("when LC_ALL is set to zh_TW", func() {
				var origLANG, origLCALL string

				BeforeEach(func() {
					origLANG = os.Getenv("LANG")
					origLCALL = os.Getenv("LC_ALL")
					os.Setenv("LANG", "")
					os.Setenv("LC_ALL", "zh_TW")
				})

				AfterEach(func() {
					os.Setenv("LANG", origLANG)
					os.Setenv("LC_ALL", origLCALL)
				})

				It("returns a translate func for zh_Hant", func() {
					translateFunc := i18n.Init(configRepo, detector)
					Expect(translateFunc("No buildpacks found")).To(Equal("(Hant)No buildpacks found"))
				})
			})

			Context("when LC_ALL is set to zh_HK", func() {
				var origLANG, origLCALL string

				BeforeEach(func() {
					origLANG = os.Getenv("LANG")
					origLCALL = os.Getenv("LC_ALL")
					os.Setenv("LANG", "")
					os.Setenv("LC_ALL", "zh_HK")
				})

				AfterEach(func() {
					os.Setenv("LANG", origLANG)
					os.Setenv("LC_ALL", origLCALL)
				})

				It("returns a translate func for zh_Hant", func() {
					translateFunc := i18n.Init(configRepo, detector)
					Expect(translateFunc("No buildpacks found")).To(Equal("(Hant)No buildpacks found"))
				})
			})

			Context("when LC_ALL is set to something other than zh_TW or zh_HK", func() {
				Context("when the language can be loaded via locale", func() {
					var origLANG, origLCALL string

					BeforeEach(func() {
						origLANG = os.Getenv("LANG")
						origLCALL = os.Getenv("LC_ALL")
						os.Setenv("LANG", "")
						os.Setenv("LC_ALL", "fr_FR")
					})

					AfterEach(func() {
						os.Setenv("LANG", origLANG)
						os.Setenv("LC_ALL", origLCALL)
					})

					It("returns a translate func for that locale", func() {
						translateFunc := i18n.Init(configRepo, detector)
						Expect(translateFunc("No buildpacks found")).To(Equal("Pas buildpacks trouvés"))
					})
				})

				Context("when the translations can be loaded via language", func() {
					var origLANG, origLCALL string

					BeforeEach(func() {
						origLANG = os.Getenv("LANG")
						origLCALL = os.Getenv("LC_ALL")
						os.Setenv("LANG", "")
						os.Setenv("LC_ALL", "fr_ZZ")
					})

					AfterEach(func() {
						os.Setenv("LANG", origLANG)
						os.Setenv("LC_ALL", origLCALL)
					})

					Context("when the translations can be loaded", func() {
						It("returns a translate func for that locale", func() {
							translateFunc := i18n.Init(configRepo, detector)
							Expect(translateFunc("No buildpacks found")).To(Equal("Pas buildpacks trouvés"))
						})
					})

					Context("when the translations cannot be loaded", func() {
						It("returns a translate func for the default locale", func() {})
					})
				})

				Context("when the translations cannot be loaded via language", func() {
					var origLANG, origLCALL string

					BeforeEach(func() {
						origLANG = os.Getenv("LANG")
						origLCALL = os.Getenv("LC_ALL")
						os.Setenv("LANG", "")
						os.Setenv("LC_ALL", "zz_ZZ")
					})

					AfterEach(func() {
						os.Setenv("LANG", origLANG)
						os.Setenv("LC_ALL", origLCALL)
					})

					It("returns a translate func for the default locale", func() {
						translateFunc := i18n.Init(configRepo, detector)
						Expect(translateFunc("No buildpacks found")).To(Equal("No buildpacks found"))
					})
				})
			})
		})
	})
})
