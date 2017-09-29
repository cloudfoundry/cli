package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"code.cloudfoundry.org/cli/i18n/resources"
	"golang.org/x/text/language"
)

const (
	// assetPath is the path of the translation file inside the asset loader.
	assetPath = "resources/%s.all.json"
	// chineseBase is the language code for Chinese.
	chineseBase = "zh"
	// defaultLocale is the default locale used when one is not configured.
	defaultLocale = "en-us"
	// unspecifiedScript is what is returned by language#Script objects when the
	// script cannot be determined.
	unspecifiedScript = "Zzzz"
)

type LocaleReader interface {
	Locale() string
}

// TranslationEntry is the expected format of the translation file.
type TranslationEntry struct {
	// ID is the original English string.
	ID string `json:"id"`
	// Translation is the translation of the ID.
	Translation string `json:"translation"`
}

// TranslateFunc returns the translation of the string identified by
// translationID.
//
// If there is no translation for translationID, then the translationID is used
// as the translation.
type TranslateFunc func(translationID string, args ...interface{}) string

// GetTranslationFunc will return back a function that can be used to translate
// strings into the currently set locale.
func GetTranslationFunc(reader LocaleReader) (TranslateFunc, error) {
	locale, err := determineLocale(reader)
	if err != nil {
		locale = defaultLocale
	}

	rawTranslation, err := loadAssetFromResources(locale)
	if err != nil {
		rawTranslation, err = loadAssetFromResources(defaultLocale)
		if err != nil {
			return nil, err
		}
	}

	return generateTranslationFunc(rawTranslation)
}

// ParseLocale will return a locale formatted as "<language code>-<region
// code>" for all non-Chinese lanagues. For Chinese, it will return
// "zh-<script>", defaulting to "hant" if script is unspecified.
func ParseLocale(locale string) (string, error) {
	lang, err := language.Parse(locale)
	if err != nil {
		return "", err
	}

	base, script, region := lang.Raw()
	switch base.String() {
	case chineseBase:
		if script.String() == unspecifiedScript {
			return "zh-hant", nil
		}
		return strings.ToLower(fmt.Sprintf("%s-%s", base, script)), nil
	default:
		return strings.ToLower(fmt.Sprintf("%s-%s", base, region)), nil
	}
}

func determineLocale(reader LocaleReader) (string, error) {
	locale := reader.Locale()
	if locale == "" {
		return defaultLocale, nil
	}

	return ParseLocale(locale)
}

func generateTranslationFunc(rawTranslation []byte) (TranslateFunc, error) {
	var entries []TranslationEntry
	err := json.Unmarshal(rawTranslation, &entries)
	if err != nil {
		return nil, err
	}

	translations := map[string]string{}
	for _, entry := range entries {
		translations[entry.ID] = entry.Translation
	}

	return func(translationID string, args ...interface{}) string {
		translated := translations[translationID]
		if translated == "" {
			translated = translationID
		}

		var keys interface{}
		if len(args) > 0 {
			keys = args[0]
		}

		var buffer bytes.Buffer
		formattedTemplate := template.Must(template.New("Display Text").Parse(translated))
		formattedTemplate.Execute(&buffer, keys)

		return buffer.String()
	}, nil
}

func loadAssetFromResources(locale string) ([]byte, error) {
	assetName := fmt.Sprintf(assetPath, locale)
	assetBytes, err := resources.Asset(assetName)
	if err != nil {
		err = fmt.Errorf("Could not load asset '%s': %s", assetName, err.Error())
	}

	return assetBytes, err
}
