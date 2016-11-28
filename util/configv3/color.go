package configv3

import "strconv"

const (
	// DefaultColorEnabled is the default CFConfig value for ColorEnabled.
	DefaultColorEnabled = ""

	// ColorDisabled means that no colors/bolding will be displayed.
	ColorDisabled ColorSetting = iota

	// ColorEnabled means colors/bolding will be displayed.
	ColorEnabled

	// ColorAuto means that the UI should decide if colors/bolding will be
	// enabled.
	ColorAuto
)

// ColorSetting is a trinary operator that represents if the display should
// have colors enabled, disabled, or automatically detected.
type ColorSetting int

// ColorEnabled returns the color setting based off:
//   1. The $CF_COLOR environment variable if set (0/1/t/f/true/false)
//   2. The 'ColorEnabled' value in the .cf/config.json if set
//   3. Defaults to ColorEnabled if nothing is set
func (config *Config) ColorEnabled() ColorSetting {
	if config.ENV.CFColor != "" {
		val, err := strconv.ParseBool(config.ENV.CFColor)
		if err == nil {
			return config.boolToColorSetting(val)
		}
	}

	val, err := strconv.ParseBool(config.ConfigFile.ColorEnabled)
	if err != nil {
		return ColorEnabled
	}
	return config.boolToColorSetting(val)
}

func (config *Config) boolToColorSetting(val bool) ColorSetting {
	if val {
		return ColorEnabled
	}

	return ColorDisabled
}
