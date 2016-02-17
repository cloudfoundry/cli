package i18n

import (
	"path"
	"strings"
	"unicode"

	"github.com/cloudfoundry/cli/cf/resources"
	"github.com/nicksnyder/go-i18n/i18n/language"
)

func SupportedLocales() []string {
	locales := supportedLocales()
	localeNames := make([]string, len(locales))

	for i, locale := range locales {
		localeNames[i] = locale.String()
	}

	return localeNames
}

func IsSupportedLocale(locale string) bool {
	for _, supportedLocale := range supportedLocales() {
		for _, l := range language.Parse(locale) {
			if supportedLocale.NormalizedString() == l.String() {
				return true
			}
		}
	}

	return false
}

func supportedLocales() []*locale {
	assetNames := resources.AssetNames()

	locales := make([]*locale, len(assetNames))

	for i, assetName := range assetNames {
		locales[i] = newLocale(assetName)
	}

	return locales
}

type locale struct {
	tag string
}

func newLocale(asset string) *locale {
	tag := strings.TrimSuffix(path.Base(asset), resourceSuffix)

	return &locale{
		tag: tag,
	}
}

func (l *locale) subtags() []string {
	return strings.Split(l.tag, "-")
}

func (l *locale) String() string {
	localeParts := l.subtags()
	language := localeParts[0]
	regionOrScript := localeParts[1]

	switch len(regionOrScript) {
	case 2: // Region
		return language + "-" + strings.ToUpper(regionOrScript)
	case 4: // Script
		return language + "-" + upcase(regionOrScript)
	}

	return l.tag
}

func (l *locale) NormalizedString() string {
	return language.NormalizeTag(l.String())
}

func upcase(s string) string {
	for i, v := range s {
		return string(unicode.ToUpper(v)) + s[i+1:]
	}
	return ""
}
