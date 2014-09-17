package main_test

import (
	"os"
	"time"

	"github.com/cloudfoundry/cli/cf/i18n"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("locales", func() {
	var oldLocale string

	BeforeEach(func() {
		oldLocale = os.Getenv("LANG")
	})

	AfterEach(func() {
		os.Setenv("LANG", oldLocale)
	})

	It("exits 0 when help is run for each language", func() {
		for _, locale := range i18n.SUPPORTED_LOCALES {
			os.Setenv("LANG", locale)
			result := Cf("help")

			Eventually(result, 3*time.Second).Should(Exit(0))
		}
	})
})
