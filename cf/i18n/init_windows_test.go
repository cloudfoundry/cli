// +build windows

package i18n_test

import (
	"os"
	"path/filepath"

	"github.com/XenoPhex/jibber_jabber"
	"github.com/cloudfoundry/cli/cf/i18n"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("i18n.Init() function", func() {
	BeforeEach(func() {
		i18n.Resources_path = filepath.Join("cf", "i18n", "test_fixtures")
		//All these tests require the system language to be English
		Ω(jibber_jabber.DetectIETF()).Should(Equal("en-US"))
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
})
