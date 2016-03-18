package gosteno

import (
	"fmt"
	"sync"
)

type L interface {
	Level() LogLevel
	Log(x LogLevel, m string, d map[string]interface{})
}

type Logger struct {
	sync.Mutex
	L
	d map[string]interface{}
}

type BaseLogger struct {
	name  string
	sinks []Sink
	level LogLevel
}

func (l *BaseLogger) Level() LogLevel {
	return l.level
}

func (l *BaseLogger) Log(x LogLevel, m string, d map[string]interface{}) {
	if l.Level().Priority < x.Priority {
		return
	}

	r := NewRecord(l.name, x, m, d)
	for _, s := range l.sinks {
		s.AddRecord(r)
		s.Flush()
	}

	if x == LOG_FATAL {
		panic(m)
	}
}

func (l *Logger) Log(x LogLevel, m string, d map[string]interface{}) {
	var r map[string]interface{}

	if d != nil && l.d != nil {
		r = make(map[string]interface{})

		// Copy the loggers data
		for k, v := range l.d {
			r[k] = v
		}

		// Overwrite specified data
		for k, v := range d {
			r[k] = v
		}
	} else if d != nil {
		r = d
	} else {
		r = l.d
	}

	l.L.Log(x, m, r)
}

func (l *Logger) Set(k string, v interface{}) {
	l.Lock()

	if l.d == nil {
		l.d = make(map[string]interface{})
	}

	l.d[k] = v

	l.Unlock()
}

func (l *Logger) Get(k string) (rv interface{}) {
	l.Lock()

	if l.d != nil {
		rv = l.d[k]
	}

	l.Unlock()

	return
}

func (l *Logger) Copy() (rv *Logger) {
	rv = &Logger{L: l.L}

	l.Lock()

	for k, v := range l.d {
		rv.Set(k, v)
	}

	l.Unlock()

	return
}

func (l *Logger) Fatal(m string) {
	l.Log(LOG_FATAL, m, nil)
}

func (l *Logger) Error(m string) {
	l.Log(LOG_ERROR, m, nil)
}

func (l *Logger) Warn(m string) {
	l.Log(LOG_WARN, m, nil)
}

func (l *Logger) Info(m string) {
	l.Log(LOG_INFO, m, nil)
}

func (l *Logger) Debug(m string) {
	l.Log(LOG_DEBUG, m, nil)
}

func (l *Logger) Debug1(m string) {
	l.Log(LOG_DEBUG1, m, nil)
}

func (l *Logger) Debug2(m string) {
	l.Log(LOG_DEBUG2, m, nil)
}

func (l *Logger) Fatald(d map[string]interface{}, m string) {
	l.Log(LOG_FATAL, m, d)
}

func (l *Logger) Errord(d map[string]interface{}, m string) {
	l.Log(LOG_ERROR, m, d)
}

func (l *Logger) Warnd(d map[string]interface{}, m string) {
	l.Log(LOG_WARN, m, d)
}

func (l *Logger) Infod(d map[string]interface{}, m string) {
	l.Log(LOG_INFO, m, d)
}

func (l *Logger) Debugd(d map[string]interface{}, m string) {
	l.Log(LOG_DEBUG, m, d)
}

func (l *Logger) Debug1d(d map[string]interface{}, m string) {
	l.Log(LOG_DEBUG1, m, d)
}

func (l *Logger) Debug2d(d map[string]interface{}, m string) {
	l.Log(LOG_DEBUG2, m, d)
}

func (l *Logger) Fatalf(f string, a ...interface{}) {
	l.Log(LOG_FATAL, fmt.Sprintf(f, a...), nil)
}

func (l *Logger) Errorf(f string, a ...interface{}) {
	l.Log(LOG_ERROR, fmt.Sprintf(f, a...), nil)
}

func (l *Logger) Warnf(f string, a ...interface{}) {
	l.Log(LOG_WARN, fmt.Sprintf(f, a...), nil)
}

func (l *Logger) Infof(f string, a ...interface{}) {
	l.Log(LOG_INFO, fmt.Sprintf(f, a...), nil)
}

func (l *Logger) Debugf(f string, a ...interface{}) {
	l.Log(LOG_DEBUG, fmt.Sprintf(f, a...), nil)
}

func (l *Logger) Debug1f(f string, a ...interface{}) {
	l.Log(LOG_DEBUG1, fmt.Sprintf(f, a...), nil)
}

func (l *Logger) Debug2f(f string, a ...interface{}) {
	l.Log(LOG_DEBUG2, fmt.Sprintf(f, a...), nil)
}

func (l *Logger) Fataldf(d map[string]interface{}, f string, a ...interface{}) {
	l.Log(LOG_FATAL, fmt.Sprintf(f, a...), d)
}

func (l *Logger) Errordf(d map[string]interface{}, f string, a ...interface{}) {
	l.Log(LOG_ERROR, fmt.Sprintf(f, a...), d)
}

func (l *Logger) Warndf(d map[string]interface{}, f string, a ...interface{}) {
	l.Log(LOG_WARN, fmt.Sprintf(f, a...), d)
}

func (l *Logger) Infodf(d map[string]interface{}, f string, a ...interface{}) {
	l.Log(LOG_INFO, fmt.Sprintf(f, a...), d)
}

func (l *Logger) Debugdf(d map[string]interface{}, f string, a ...interface{}) {
	l.Log(LOG_DEBUG, fmt.Sprintf(f, a...), d)
}

func (l *Logger) Debug1df(d map[string]interface{}, f string, a ...interface{}) {
	l.Log(LOG_DEBUG1, fmt.Sprintf(f, a...), d)
}

func (l *Logger) Debug2df(d map[string]interface{}, f string, a ...interface{}) {
	l.Log(LOG_DEBUG2, fmt.Sprintf(f, a...), d)
}
