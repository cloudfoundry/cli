package gosteno

import (
	"regexp"
)

// loggerRegexp* used to match log name and log level
var loggerRegexp *regexp.Regexp
var loggerRegexpLevel *LogLevel

func SetLoggerRegexp(pattern string, level LogLevel) error {
	loggersMutex.Lock()
	defer loggersMutex.Unlock()

	clearLoggerRegexp()
	return setLoggerRegexp(pattern, level)
}

func ClearLoggerRegexp() {
	loggersMutex.Lock()
	defer loggersMutex.Unlock()

	clearLoggerRegexp()
}

func setLoggerRegexp(pattern string, level LogLevel) error {
	regExp, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	loggerRegexp = regExp
	loggerRegexpLevel = &level

	for name, logger := range loggers {
		if loggerRegexp.MatchString(name) {
			logger.level = level
		}
	}

	return nil
}

func clearLoggerRegexp() {
	if loggerRegexp == nil {
		return
	}

	for name, logger := range loggers {
		if loggerRegexp.MatchString(name) {
			logger.level = getConfig().Level
		}
	}

	loggerRegexp = nil
	loggerRegexpLevel = nil
}

func computeLevel(name string) LogLevel {
	if loggerRegexpLevel != nil && loggerRegexp.MatchString(name) {
		return *loggerRegexpLevel
	}

	return getConfig().Level
}
