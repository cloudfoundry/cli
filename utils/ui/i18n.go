package ui

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"

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

// GetTranslationFunc will return back a function that can be used to translate
// strings into the currently set locale.
func GetTranslationFunc(config Config) (i18n.TranslateFunc, error) {
	t, err := getConfiguredLocal(config)
	if err != nil {
		return nil, err
	}

	if t == nil {
		t, err = getDefaultLocal()
		if err != nil {
			return nil, err
		}
	}

	return translationWrapper(t), nil
}

func translationWrapper(translationFunc i18n.TranslateFunc) i18n.TranslateFunc {
	return func(translationID string, args ...interface{}) string {
		var keys interface{}
		if len(args) > 0 {
			keys = args[0]
		}

		if translated := translationFunc(translationID, keys); translated != translationID {
			return translated
		}

		var buffer bytes.Buffer
		formattedTemplate := template.Must(template.New("Display Text").Parse(translationID))
		formattedTemplate.Execute(&buffer, keys)

		return buffer.String()
	}
}

func getConfiguredLocal(config Config) (i18n.TranslateFunc, error) {
	source := config.Locale()
	assetNames := resources.AssetNames()

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

				return i18n.MustTfunc(l.Tag), nil
			}
		}
	}

	return nil, nil
}

func getDefaultLocal() (i18n.TranslateFunc, error) {
	err := loadAsset("cf/i18n/resources/" + defaultLocale + resourceSuffix)
	if err != nil {
		return nil, err
	}

	return i18n.MustTfunc(defaultLocale), nil
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
