package i18n

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	resources "github.com/cloudfoundry/cli/cf/resources"
	go_i18n "github.com/nicksnyder/go-i18n/i18n"
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
	for _, assetName := range resources.AssetNames() {
		assetBytes, err := resources.Asset(assetName)
		if err != nil {
			panic(fmt.Sprintf("Could not load asset '%s': %s", assetName, err.Error()))
		}
		err = go_i18n.ParseTranslationFileBytes(assetName, assetBytes)
		if err != nil {
			panic(fmt.Sprintf("Could not load translations '%s': %s", assetName, err.Error()))
		}
	}

	T, err := go_i18n.Tfunc(config.Locale(), os.Getenv("LC_ALL"), os.Getenv("LANG"), DEFAULT_LOCALE)
	if err != nil {
		panic(fmt.Sprintf("Failed to create translation func", err.Error()))
	}

	return T
}
