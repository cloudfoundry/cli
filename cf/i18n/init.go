package i18n

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf-experimental/jibber_jabber"

	resources "github.com/cloudfoundry/cli/cf/resources"
	go_i18n "github.com/nicksnyder/go-i18n/i18n"
)

const (
	DEFAULT_LOCALE   = "en_US"
	DEFAULT_LANGUAGE = "en"
)

var SUPPORTED_LANGUAGES = []string{"en", "fr"}
var SUPPORTED_LOCALES = map[string]string{"en": "en_US", "fr": "fr_FR"}
var resources_path = filepath.Join("cf", "i18n", "resources")

func GetResourcesPath() string {
	return resources_path
}

func Init(packageName string, i18nDirname string) go_i18n.TranslateFunc {
	userLocale, err := initWithUserLocale(packageName, i18nDirname)
	if err != nil {
		userLocale = mustLoadDefaultLocale(packageName, i18nDirname)
	}

	T, err := go_i18n.Tfunc(userLocale, DEFAULT_LOCALE)
	if err != nil {
		panic(err)
	}

	return T
}

func initWithUserLocale(packageName, i18nDirname string) (string, error) {
	userLocale, err := jibber_jabber.DetectIETF()
	if err != nil {
		userLocale = DEFAULT_LOCALE
	}

	language, err := jibber_jabber.DetectLanguage()
	if err != nil {
		language = DEFAULT_LANGUAGE
	}

	userLocale = strings.Replace(userLocale, "-", "_", 1)
	err = loadFromAsset(packageName, i18nDirname, userLocale, language)
	if err != nil {
		locale := SUPPORTED_LOCALES[language]
		if locale == "" {
			userLocale = DEFAULT_LOCALE
		} else {
			userLocale = locale
		}
		err = loadFromAsset(packageName, i18nDirname, userLocale, language)
	}

	return userLocale, err
}

func mustLoadDefaultLocale(packageName, i18nDirname string) string {
	userLocale := DEFAULT_LOCALE

	err := loadFromAsset(packageName, i18nDirname, DEFAULT_LOCALE, DEFAULT_LANGUAGE)
	if err != nil {
		panic("Could not load en_US language files. God save the queen. " + err.Error())
	}

	return userLocale
}

func loadFromAsset(packageName, assetPath, locale, language string) error {
	assetName := locale + ".all.json"
	assetKey := filepath.Join(assetPath, language, packageName, assetName)

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
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	fileName, err := saveLanguageFileToDisk(tmpDir, assetName, byteArray)
	if err != nil {
		return err
	}

	go_i18n.MustLoadTranslationFile(fileName)

	os.RemoveAll(fileName)

	return nil
}

func saveLanguageFileToDisk(tmpDir, assetName string, byteArray []byte) (fileName string, err error) {
	fileName = filepath.Join(tmpDir, assetName)
	file, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.Write(byteArray)
	if err != nil {
		return
	}

	return
}
