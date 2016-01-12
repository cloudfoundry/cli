package i18n

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/resources"
	go_i18n "github.com/nicksnyder/go-i18n/i18n"
	"github.com/nicksnyder/go-i18n/i18n/language"
)

const (
	defaultLocale  = "en_US"
	lang           = "LANG"
	lcAll          = "LC_ALL"
	resourceSuffix = ".all.json"
	zhTW           = "zh-tw"
	zhHK           = "zh-hk"
	zhHant         = "zhHant"
	hyphen         = "-"
	underscore     = "_"
)

var T go_i18n.TranslateFunc

func Init(config core_config.Reader) go_i18n.TranslateFunc {
	loadAsset("cf/i18n/resources/" + defaultLocale + resourceSuffix)
	defaultTfunc := go_i18n.MustTfunc(defaultLocale)

	assetNames := resources.AssetNames()

	sources := []string{
		config.Locale(),
		os.Getenv(lcAll),
		os.Getenv(lang),
	}

	for _, source := range sources {
		if source == "" {
			continue
		}

		for _, l := range language.Parse(source) {
			if l.Tag == zhTW || l.Tag == zhHK {
				l.Tag = zhHant
			}

			for _, assetName := range assetNames {
				assetLocale := strings.ToLower(strings.Replace(path.Base(assetName), underscore, hyphen, -1))
				if strings.HasPrefix(assetLocale, l.Tag) {
					loadAsset(assetName)

					t := go_i18n.MustTfunc(source)

					return func(translationID string, args ...interface{}) string {
						if translated := t(translationID, args...); translated != translationID {
							return translated
						}

						return defaultTfunc(translationID, args...)
					}
				}
			}
		}
	}

	return defaultTfunc
}

func loadAsset(assetName string) {
	assetBytes, err := resources.Asset(assetName)
	if err != nil {
		panic(fmt.Sprintf("Could not load asset '%s': %s", assetName, err.Error()))
	}

	err = go_i18n.ParseTranslationFileBytes(assetName, assetBytes)
	if err != nil {
		panic(fmt.Sprintf("Could not load translations '%s': %s", assetName, err.Error()))
	}
}

func SupportedLocales() []string {
	assetNames := resources.AssetNames()
	locales := make([]string, len(assetNames))

	for i := range assetNames {
		locales[i] = strings.TrimSuffix(path.Base(assetNames[i]), resourceSuffix)
	}

	return locales
}

func NormalizedSupportedLocales() []string {
	supportedLocales := SupportedLocales()
	locales := make([]string, len(supportedLocales))

	for i := range supportedLocales {
		locales[i] = language.NormalizeTag(supportedLocales[i])
	}

	return locales
}

func IsSupportedLocale(locale string) bool {
	for _, supportedLocale := range NormalizedSupportedLocales() {
		for _, l := range language.Parse(locale) {
			if supportedLocale == l.String() {
				return true
			}
		}
	}

	return false
}
