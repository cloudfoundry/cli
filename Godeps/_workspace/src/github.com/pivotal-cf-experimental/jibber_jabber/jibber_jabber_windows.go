// +build windows

package jibber_jabber

import (
	"errors"
	"syscall"
	"unsafe"
)

const LOCALE_NAME_MAX_LENGTH uint32 = 85

func getWindowsLocaleFrom(sysCall string) (locale string, err error) {
	buffer := make([]uint16, LOCALE_NAME_MAX_LENGTH)

	dll := syscall.MustLoadDLL("kernel32")
	proc := dll.MustFindProc(sysCall)
	r, _, dllError := proc.Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(LOCALE_NAME_MAX_LENGTH))
	if r == 0 {
		err = errors.New(COULD_NOT_DETECT_PACKAGE_ERROR_MESSAGE + ":\n" + dllError.Error())
		return
	}

	locale = syscall.UTF16ToString(buffer)

	return
}

func getWindowsLocale() (locale string, err error) {
	locale, err = getWindowsLocaleFrom("GetUserDefaultLocaleName")
	if err != nil {
		locale, err = getWindowsLocaleFrom("GetSystemDefaultLocaleName")
	}
	return
}
func DetectIETF() (locale string, err error) {
	locale, err = getWindowsLocale()
	return
}

func DetectLanguage() (language string, err error) {
	windows_locale, err := getWindowsLocale()
	if err == nil {
		language, _ = splitLocale(windows_locale)
	}

	return
}

func DetectTerritory() (territory string, err error) {
	windows_locale, err := getWindowsLocale()
	if err == nil {
		_, territory = splitLocale(windows_locale)
	}

	return
}
