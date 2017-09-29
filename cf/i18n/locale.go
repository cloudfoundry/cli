package i18n

import (
	"path"
	"strings"

	"code.cloudfoundry.org/cli/i18n/resources"
	"code.cloudfoundry.org/cli/util/ui"
)

const resourceSuffix = ".all.json"

func SupportedLocales() []string {
	languages := supportedLanguages()
	localeNames := make([]string, len(languages))

	for i, l := range languages {
		localeParts := strings.Split(l, "-")
		lang := localeParts[0]
		regionOrScript := localeParts[1]

		switch len(regionOrScript) {
		case 2: // Region
			localeNames[i] = lang + "-" + strings.ToUpper(regionOrScript)
		case 4: // Script
			localeNames[i] = lang + "-" + strings.Title(regionOrScript)
		default:
			localeNames[i] = l
		}
	}

	return localeNames
}

func IsSupportedLocale(locale string) bool {
	sanitizedLocale, err := ui.ParseLocale(locale)
	if err != nil {
		return false
	}

	for _, supportedLanguage := range supportedLanguages() {
		if supportedLanguage == sanitizedLocale {
			return true
		}
	}

	return false
}

func supportedLanguages() []string {
	assetNames := resources.AssetNames()

	var languages []string
	for _, assetName := range assetNames {
		assetLocale := strings.TrimSuffix(path.Base(assetName), resourceSuffix)
		locale, _ := ui.ParseLocale(assetLocale)
		languages = append(languages, locale)
	}

	return languages
}
