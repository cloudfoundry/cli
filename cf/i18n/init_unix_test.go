// +build darwin freebsd linux netbsd openbsd

package i18n_test

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/i18n"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("i18n.Init() function", func() {
	var oldResourcesPath string

	BeforeEach(func() {
		oldResourcesPath = i18n.GetResourcesPath()
		i18n.Resources_path = filepath.Join("cf", "i18n", "test_fixtures")
	})

	AfterEach(func() {
		i18n.Resources_path = oldResourcesPath
		os.Setenv("LC_ALL", "")
		os.Setenv("LANG", "en_US.UTF-8")
	})

	Context("loads correct locale", func() {
		It("defaults to en_US when LC_ALL and LANG not set", func() {
			os.Setenv("LC_ALL", "")
			os.Setenv("LANG", "")

			T := i18n.Init()

			translation := T("Hello world!")
			Ω("Hello world!").Should(Equal(translation))
		})

		Context("when there is no territory set", func() {
			BeforeEach(func() {
				os.Setenv("LANG", "en")
			})

			It("still loads the english translation", func() {
				T := i18n.Init()

				translation := T("Hello world!")
				Ω("Hello world!").Should(Equal(translation))
			})
		})

		Context("when the desired language is not supported", func() {
			It("defaults to en_US when langauge is not supported", func() {
				os.Setenv("LC_ALL", "zz_FF.UTF-8")
				T := i18n.Init()

				translation := T("Hello world!")
				Ω("Hello world!").Should(Equal(translation))

				translation = T("Hello {{.Adj}} new world!", map[string]interface{}{"Adj": "brave"})
				Ω("Hello brave new world!").Should(Equal(translation))
			})

			Context("because we don't have the territory", func() {
				It("defaults to same language in supported territory", func() {
					os.Setenv("LC_ALL", "fr_CA.UTF-8")
					T := i18n.Init()

					translation := T("Hello world!")
					Ω("Àlo le monde!").Should(Equal(translation))
				})
			})
		})

		Context("when not even the english translation can be loaded", func() {
			BeforeEach(func() {
				i18n.Resources_path = filepath.Join("should", "not", "be_valid")
			})

			It("panics", func() {
				os.Setenv("LC_ALL", "zz_FF.utf-8")

				init := func() { i18n.Init() }
				Ω(init).Should(Panic(), "loading translations from an invalid path should panic")
			})
		})
	})

	Context("creates a valid T function", func() {
		BeforeEach(func() {
			os.Setenv("LC_ALL", "en_US.UTF-8")
		})

		It("returns a usable T function for simple strings", func() {
			T := i18n.Init()
			Ω(T).ShouldNot(BeNil())

			translation := T("Hello world!")
			Ω("Hello world!").Should(Equal(translation))
		})

		It("returns a usable T function for complex strings (interpolated)", func() {
			T := i18n.Init()
			Ω(T).ShouldNot(BeNil())

			translation := T("Hello {{.Name}}!", map[string]interface{}{"Name": "Anand"})
			Ω("Hello Anand!").Should(Equal(translation))
		})
	})

	Context("translates correctly", func() {
		BeforeEach(func() {
			os.Setenv("LC_ALL", "fr_FR.UTF-8")
		})

		It("T function should return translation if string key exists", func() {
			T := i18n.Init()

			translation := T("Hello world!")
			Ω("Àlo le monde!").Should(Equal(translation))
		})
	})
})
