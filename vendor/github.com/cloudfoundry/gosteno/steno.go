package gosteno

import (
	"encoding/json"
	"sync"
)

// Global configs
var config Config

// loggersMutex protects accesses to loggers and regexp
var loggersMutex = &sync.Mutex{}

// loggersMutex protects accesses to loggers and regexp
var configMutex = &sync.RWMutex{}

// loggers only saves BaseLogger
var loggers = make(map[string]*BaseLogger)

func Init(c *Config) {
	loggersMutex.Lock()
	defer loggersMutex.Unlock()

	if c.Level == (LogLevel{}) {
		c.Level = LOG_INFO
	}

	if c.Codec == nil {
		c.Codec = NewJsonCodec()
	}

	if c.Sinks == nil {
		c.Sinks = []Sink{}
	}

	for _, sink := range c.Sinks {
		if sink.GetCodec() == nil {
			sink.SetCodec(c.Codec)
		}
	}

	setConfig(*c)

	for name, _ := range loggers {
		loggers[name] = nil
	}
}

func NewLogger(name string) *Logger {
	loggersMutex.Lock()
	defer loggersMutex.Unlock()

	l := loggers[name]
	if l == nil {
		bl := &BaseLogger{
			name:  name,
			sinks: getConfig().Sinks,
			level: computeLevel(name),
		}

		loggers[name] = bl
		l = bl
	}

	return &Logger{L: l}
}

func getConfig() Config {
	configMutex.RLock()
	defer configMutex.RUnlock()

	return config
}

func setConfig(newConfig Config) {
	configMutex.Lock()
	defer configMutex.Unlock()

	config = newConfig
}

func loggersInJson() string {
	bytes, _ := json.Marshal(loggers)
	return string(bytes)
}
