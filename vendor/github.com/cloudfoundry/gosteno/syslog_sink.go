// +build !windows,!plan9

package gosteno

import (
	"errors"
	syslog "github.com/cloudfoundry/gosteno/syslog"
	"sync"
)

const (
	MaxMessageSize  = 1024 * 3
	TruncatePostfix = "..."
)

type Syslog struct {
	writer *syslog.Writer
	codec  Codec

	sync.Mutex
}

func NewSyslogSink(namespace string) *Syslog {
	writer, err := syslog.New(syslog.LOG_DEBUG, namespace)
	if err != nil {
		panic(errors.New("Could not setup logging to syslog: " + err.Error()))
	}

	syslog := new(Syslog)
	syslog.writer = writer
	return syslog
}

func (s *Syslog) AddRecord(record *Record) {
	truncate(record)

	bytes, _ := s.codec.EncodeRecord(record)
	msg := string(bytes)

	s.Lock()
	defer s.Unlock()

	switch record.Level {
	case LOG_FATAL:
		s.writer.Crit(msg)
	case LOG_ERROR:
		s.writer.Err(msg)
	case LOG_WARN:
		s.writer.Warning(msg)
	case LOG_INFO:
		s.writer.Info(msg)
	case LOG_DEBUG, LOG_DEBUG1, LOG_DEBUG2:
		s.writer.Debug(msg)
	default:
		panic("Unknown log level: " + record.Level.Name)
	}
}

func (s *Syslog) Flush() {
	// No impl.
}

func (s *Syslog) SetCodec(codec Codec) {
	s.Lock()
	defer s.Unlock()

	s.codec = codec
}

func (s *Syslog) GetCodec() Codec {
	s.Lock()
	defer s.Unlock()

	return s.codec
}

func truncate(record *Record) {
	if len(record.Message) <= MaxMessageSize {
		return
	}

	record.Message = record.Message[:MaxMessageSize-len(TruncatePostfix)] + TruncatePostfix
}
