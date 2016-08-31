package ui

import (
	"fmt"
	"path"
	"strings"

	"code.cloudfoundry.org/cli/cf/resources"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/nicksnyder/go-i18n/i18n/language"
)

const (
	defaultLocale  = "en-us"
	resourceSuffix = ".all.json"
	zhTW           = "zh-tw"
	zhHK           = "zh-hk"
	zhHant         = "zh-hant"
	hyphen         = "-"
	underscore     = "_"
)

func GetTranslationFunc(config Config) (i18n.TranslateFunc, error) {
	err := loadAsset("cf/i18n/resources/" + defaultLocale + resourceSuffix)
	if err != nil {
		return nil, err
	}
	defaultTfunc := i18n.MustTfunc(defaultLocale)

	assetNames := resources.AssetNames()

	source := config.Locale()
	for _, l := range language.Parse(source) {
		if l.Tag == zhTW || l.Tag == zhHK {
			l.Tag = zhHant
		}

		for _, assetName := range assetNames {
			assetLocale := strings.ToLower(strings.Replace(path.Base(assetName), underscore, hyphen, -1))
			if strings.HasPrefix(assetLocale, l.Tag) {
				err := loadAsset(assetName)
				if err != nil {
					return nil, err
				}

				t := i18n.MustTfunc(l.Tag)

				return func(translationID string, args ...interface{}) string {
					if translated := t(translationID, args...); translated != translationID {
						return translated
					}

					return defaultTfunc(translationID, args...)
				}, nil
			}
		}
	}

	return defaultTfunc, nil
}

func loadAsset(assetName string) error {
	assetBytes, err := resources.Asset(assetName)
	if err != nil {
		return fmt.Errorf("Could not load asset '%s': %s", assetName, err.Error())
	}

	err = i18n.ParseTranslationFileBytes(assetName, assetBytes)
	if err != nil {
		return fmt.Errorf("Could not load translations '%s': %s", assetName, err.Error())
	}
	return nil
}
