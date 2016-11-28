package configv3

import "strings"

const (
	// DefaultLocale is the default CFConfig value for Locale.
	DefaultLocale = ""
)

// Locale returns the locale/language the UI should be displayed in. This value
// is based off of:
//   1. The 'Locale' setting in the .cf/config.json
//   2. The $LC_ALL environment variable if set
//   3. The $LANG environment variable if set
//   4. Defaults to DefaultLocale
func (config *Config) Locale() string {
	if config.ConfigFile.Locale != "" {
		return config.ConfigFile.Locale
	}

	if config.ENV.LCAll != "" {
		return config.convertLocale(config.ENV.LCAll)
	}

	if config.ENV.Lang != "" {
		return config.convertLocale(config.ENV.Lang)
	}

	return DefaultLocale
}

func (config *Config) convertLocale(local string) string {
	lang := strings.Split(local, ".")[0]
	return strings.Replace(lang, "_", "-", -1)
}
