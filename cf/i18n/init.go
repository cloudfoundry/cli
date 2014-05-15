package i18n

import (
	"fmt"
	"os"
	"path/filepath"

	go_i18n "github.com/nicksnyder/go-i18n/i18n"
)

func Init(packageName string, i18nDirname string) (go_i18n.TranslateFunc, error) {
	osLocale := os.Getenv("LANG")

	userLocale := osLocale[:len(".UTF8")] //Might need to make this generic
	defaultLocale := "en"

	go_i18n.MustLoadTranslationFile(filepath.Join(i18nDirname, packageName, defaultLocale) + ".all.json")
	go_i18n.MustLoadTranslationFile(filepath.Join(i18nDirname, packageName, userLocale) + ".all.json")

	T, err := go_i18n.Tfunc(userLocale, defaultLocale)
	if err != nil {
		fmt.Printf("Could not initialize i18n strings")
		return nil, err
	}

	return T, nil
}
