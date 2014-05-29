package jibber_jabber

import (
	"strings"
)

const (
	COULD_NOT_DETECT_PACKAGE_ERROR_MESSAGE = "Could not detect Language"
)

func splitLocale(locale string) (string, string) {
	formattedLocale := strings.Split(locale, ".")[0]
	formattedLocale = strings.Replace(formattedLocale, "-", "_", -1)
	language := strings.Split(formattedLocale, "_")[0]
	territory := strings.Split(formattedLocale, "_")[1]
	return language, territory
}
