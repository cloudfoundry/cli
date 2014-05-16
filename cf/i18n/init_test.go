package i18n_test

import (
	"github.com/cloudfoundry/cli/cf/i18n"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("i18n.Init() function", func() {
	var I18N_PATH = filepath.Join("cf", "i18n", "test_fixtures")

	Context("loads correct local", func() {
		It("selects LC_ALL when set", func() {
			os.Setenv("LC_ALL", "fr_FR.UTF-8")

			T, err := i18n.Init("main", I18N_PATH)
			Ω(err).ShouldNot(HaveOccurred())

			translation := T("Hello world!")
			Ω("Àlo le monde!").Should(Equal(translation))
		})

		It("selects LANG when LC_ALL not set", func() {
			os.Setenv("LC_ALL", "")
			os.Setenv("LANG", "fr_FR.UTF-8")

			T, err := i18n.Init("main", I18N_PATH)
			Ω(err).ShouldNot(HaveOccurred())

			translation := T("Hello world!")
			Ω("Àlo le monde!").Should(Equal(translation))
		})

		It("defaults to en_US when LC_ALL and LANG not set", func() {
			os.Setenv("LC_ALL", "")
			os.Setenv("LANG", "")

			T, err := i18n.Init("main", I18N_PATH)
			Ω(err).ShouldNot(HaveOccurred())

			translation := T("Hello world!")
			Ω("Hello world!").Should(Equal(translation))
		})

		It("defaults to en_US when langauge is not supported", func() {
			os.Setenv("LC_ALL", "zz_FF.UTF-8")

			T, err := i18n.Init("main", I18N_PATH)
			Ω(err).ShouldNot(HaveOccurred())

			translation := T("Hello world!")
			Ω("Hello world!").Should(Equal(translation))
		})
	})

	Context("creates a valid T function", func() {
		BeforeEach(func() {
			os.Setenv("LC_ALL", "en_US.UTF-8")
		})

		It("returns a usable T function", func() {
			T, err := i18n.Init("main", I18N_PATH)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(T).ShouldNot(BeNil())

			translation := T("Hello world!")
			Ω("Hello world!").Should(Equal(translation))
		})
	})

	Context("translates correctly", func() {
		It("T function should return translation if string key exists", func() {
			os.Setenv("LC_ALL", "fr_FR.UTF-8")

			T, err := i18n.Init("main", I18N_PATH)
			Ω(err).ShouldNot(HaveOccurred())

			translation := T("Hello world!")
			Ω("Àlo le monde!").Should(Equal(translation))
		})

		It("T function should return translation if it exists", func() {
			os.Setenv("LC_ALL", "fr_FR.UTF-8")

			T, err := i18n.Init("main", I18N_PATH)
			Ω(err).ShouldNot(HaveOccurred())

			translation := T("NSFW")
			Ω("NSFW").Should(Equal(translation))
		})

	})

	Context("formats locale correctly to xx_YY", func() {
		It("remove dash to underscore", func() {
			os.Setenv("LC_ALL", "fr-FR.UTF-8")

			T, err := i18n.Init("main", I18N_PATH)
			Ω(err).ShouldNot(HaveOccurred())

			translation := T("Hello world!")
			Ω("Àlo le monde!").Should(Equal(translation))
		})

		It("correcting language", func() {
			os.Setenv("LC_ALL", "EN_US.UTF-8")

			T, err := i18n.Init("main", I18N_PATH)
			Ω(err).ShouldNot(HaveOccurred())

			translation := T("Hello world!")
			Ω("Hello world!").Should(Equal(translation))
		})

		It("correcting teritorry", func() {
			os.Setenv("LC_ALL", "en_us.UTF-8")

			T, err := i18n.Init("main", I18N_PATH)
			Ω(err).ShouldNot(HaveOccurred())

			translation := T("Hello world!")
			Ω("Hello world!").Should(Equal(translation))
		})
	})

	Context("Loading nonexistant asset", func() {
		It("should fail", func() {
			os.Setenv("LC_ALL", "fr_FR.UTF-8")

			_, err := i18n.Init("lol", I18N_PATH)
			Ω(err).Should(HaveOccurred())
		})
	})
})
