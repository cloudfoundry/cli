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

const DEFAULT_LOCALE = "en_US"

var T go_i18n.TranslateFunc

var SUPPORTED_LOCALES = map[string]string{
	"de": "de_DE",
	"en": "en_US",
	"es": "es_ES",
	"fr": "fr_FR",
	"it": "it_IT",
	"ja": "ja_JA",
	"ko": "ko_KR",
	"pt": "pt_BR",
	//"ru": "ru_RU", - Will add support for Russian when nicksnyder/go-i18n supports Russian
	"zh": "zh_Hans",
}

func Init(config core_config.Reader) go_i18n.TranslateFunc {
	sources := []string{
		config.Locale(),
		os.Getenv("LC_ALL"),
		os.Getenv("LANG"),
		DEFAULT_LOCALE,
	}

	assetNames := resources.AssetNames()

	for _, source := range sources {
		if source == "" {
			continue
		}

		for _, l := range language.Parse(source) {
			if l.Tag == "zh-tw" || l.Tag == "zh-hk" {
				l.Tag = "zh-hant"
			}

			for _, assetName := range assetNames {
				assetLocale := strings.ToLower(strings.Replace(path.Base(assetName), "_", "-", -1))
				if strings.HasPrefix(assetLocale, l.Tag) {
					assetBytes, err := resources.Asset(assetName)
					if err != nil {
						panic(fmt.Sprintf("Could not load asset '%s': %s", assetName, err.Error()))
					}

					err = go_i18n.ParseTranslationFileBytes(assetName, assetBytes)
					if err != nil {
						panic(fmt.Sprintf("Could not load translations '%s': %s", assetName, err.Error()))
					}

					T, err := go_i18n.Tfunc(source)
					if err == nil {
						return T
					}
				}
			}
		}
	}

	panic("Unable to find suitable translation")
}
