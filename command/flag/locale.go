package flag

import (
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/cf/i18n"
	flags "github.com/jessevdk/go-flags"
)

type Locale struct {
	Locale string
}

func (l Locale) Complete(prefix string) []flags.Completion {
	return completions(l.listLocales(), l.sanitize(prefix), false)
}

func (l *Locale) UnmarshalFlag(val string) error {
	sanitized := strings.ToLower(l.sanitize(val))
	for _, locale := range l.listLocales() {
		if sanitized == strings.ToLower(locale) {
			l.Locale = locale
			return nil
		}
	}

	return &flags.Error{
		Type:    flags.ErrRequired,
		Message: fmt.Sprintf("LOCALE must be %s", strings.Join(l.listLocales(), ", ")),
	}
}

func (Locale) sanitize(val string) string {
	return strings.Replace(val, "_", "-", -1)
}

func (Locale) listLocales() []string {
	locals := append(i18n.SupportedLocales(), "CLEAR")
	sort.Strings(locals)
	return locals
}
