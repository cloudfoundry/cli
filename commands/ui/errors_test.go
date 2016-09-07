package ui_test

import (
	. "code.cloudfoundry.org/cli/commands/ui"

	"github.com/nicksnyder/go-i18n/i18n"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errors", func() {
	It("expects all errors conform to the TranslatableError interface", func() {
		errors := []TranslatableError{
			APIRequestError{},
			InvalidSSLCertError{},
		}

		for _, err := range errors {
			translatedErr := err.SetTranslation(i18n.IdentityTfunc())
			Expect(func() { translatedErr.Error() }).ToNot(Panic())
		}
	})
})
