package i18n

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	resources "github.com/cloudfoundry/cli/cf/resources"

	go_i18n "github.com/nicksnyder/go-i18n/i18n"
)

const (
	DEFAULT_LOCAL = "en_US"
)

var SUPPORTED_LANGUAGES = []string{"ar", "ca", "zh", "cs", "da", "nl", "en", "fr", "de", "it", "ja", "lt", "pt", "es"}

func Init(packageName string, i18nDirname string) (go_i18n.TranslateFunc, error) {
	userLocale := getUserLocale()
	assetPath := filepath.Join(i18nDirname, packageName)
	err := loadTranslationAssets(assetPath, userLocale, DEFAULT_LOCAL)
	if err != nil {
		return nil, err
	}

	T, err := go_i18n.Tfunc(userLocale, DEFAULT_LOCAL)
	if err != nil {
		fmt.Printf("Could not initialize i18n strings") //TODO: Better Error Handling
		return nil, err
	}

	return T, nil
}

func ValidateLocale(locale string) bool {
	language, territory := splitLocale(locale)
	return ValidateLanguage(language) && ValidateTerritory(territory)
}

func ValidateLanguage(language string) bool {
	for _, lang := range SUPPORTED_LANGUAGES {
		if language == lang {
			return true
		}
	}

	return false
}

func ValidateTerritory(territory string) bool {
	//TODO: complete me!
	return true
}

func getOSLocale() string {
	if os.Getenv("LC_ALL") != "" {
		return os.Getenv("LC_ALL")
	} else if os.Getenv("LANG") != "" {
		return os.Getenv("LANG")
	} else {
		return DEFAULT_LOCAL
	}
}

func formatLocale(locale string) string {
	language, territory := splitLocale(locale)
	return strings.ToLower(language) + "_" + strings.ToUpper(territory)
}

func splitLocale(locale string) (string, string) {
	formattedLocale := strings.Split(locale, ".")[0]
	formattedLocale = strings.Replace(formattedLocale, "-", "_", -1)
	language := strings.Split(formattedLocale, "_")[0]
	territory := strings.Split(formattedLocale, "_")[1]
	return language, territory
}

func getUserLocale() string {
	osLocale := getOSLocale()
	osLocale = formatLocale(osLocale)

	if !ValidateLocale(osLocale) {
		osLocale = DEFAULT_LOCAL
	}

	return osLocale
}

func loadTranslationFiles(packageName, i18nDirname, userLocale, defaultLocale string) error {
	pwd := os.Getenv("PWD")
	go_i18n.MustLoadTranslationFile(filepath.Join(pwd, i18nDirname, packageName, userLocale) + ".all.json")
	go_i18n.MustLoadTranslationFile(filepath.Join(pwd, i18nDirname, packageName, defaultLocale) + ".all.json")

	return nil
}

func loadTranslationAssets(assetPath, userLocale, defaultLocale string) error {
	err := loadFromAsset(assetPath, userLocale)
	if err != nil {
		return err
	}

	err = loadFromAsset(assetPath, defaultLocale)
	if err != nil {
		return err
	}

	return nil
}

func loadFromAsset(assetPath, locale string) error {
	assetName := locale + ".all.json"
	assetKey := filepath.Join(assetPath, assetName)

	byteArray, err := resources.Asset(assetKey)
	if err != nil {
		return err
	}

	if len(byteArray) == 0 {
		return errors.New(fmt.Sprintf("Could not load i18n asset: %v", assetKey))
	}

	tmpDir, err := ioutil.TempDir("", "cloudfoundry_cli_i18n_res")
	if err != nil {
		return err
	}

	fileName := filepath.Join(tmpDir, assetName)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	_, err = file.Write(byteArray)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	go_i18n.MustLoadTranslationFile(fileName)

	os.RemoveAll(fileName)

	return nil
}
