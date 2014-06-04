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
	var I18N_PATH = filepath.Join("cf", "i18n", "test_fixtures")

	AfterEach(func() {
		os.Setenv("LC_ALL", "")
		os.Setenv("LANG", "en_US.UTF-8")
	})

	Context("loads correct locale", func() {
		It("defaults to en_US when LC_ALL and LANG not set", func() {
			os.Setenv("LC_ALL", "")
			os.Setenv("LANG", "")

			T := i18n.Init("main", I18N_PATH)

			translation := T("Hello world!")
			Ω("Hello world!").Should(Equal(translation))
		})

		Context("when there is no territory set", func() {
			BeforeEach(func() {
				os.Setenv("LANG", "en")
			})

			It("still loads the english translation", func() {
				T := i18n.Init("main", I18N_PATH)

				translation := T("Hello world!")
				Ω("Hello world!").Should(Equal(translation))
			})
		})

		Context("when the desired language is not supported", func() {
			It("defaults to en_US when langauge is not supported", func() {
				os.Setenv("LC_ALL", "zz_FF.UTF-8")
				T := i18n.Init("main", I18N_PATH)

				translation := T("Hello world!")
				Ω("Hello world!").Should(Equal(translation))
			})
		})
	})

	Context("creates a valid T function", func() {
		BeforeEach(func() {
			os.Setenv("LC_ALL", "en_US.UTF-8")
		})

		It("returns a usable T function for simple strings", func() {
			T := i18n.Init("main", I18N_PATH)
			Ω(T).ShouldNot(BeNil())

			translation := T("Hello world!")
			Ω("Hello world!").Should(Equal(translation))
		})

		It("returns a usable T function for complex strings (interpolated)", func() {
			T := i18n.Init("main", I18N_PATH)
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
			T := i18n.Init("main", I18N_PATH)

			translation := T("Hello world!")
			Ω("Àlo le monde!").Should(Equal(translation))
		})
	})
})
