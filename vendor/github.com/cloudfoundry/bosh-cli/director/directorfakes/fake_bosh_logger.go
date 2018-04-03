package directorfakes

import (
	"fmt"
	"time"
)

type Logger struct {
	LogCallArgs *[]LogCallArgs
}

type LogCallArgs struct {
	LogLevel string
	Tag      string
	Msg      string
	Args     []string
}

func NewFakeLogger(logCalls *[]LogCallArgs) Logger {
	return Logger{
		LogCallArgs: logCalls,
	}
}

func (l Logger) castToString(args []interface{}) []string {
	b := make([]string, len(args))
	for i := range args {
		b[i] = fmt.Sprintf("%s", args[i])
	}
	return b
}

func (l Logger) logMsg(logLevel, tag, msg string, args []interface{}) {
	*l.LogCallArgs = append(*l.LogCallArgs, LogCallArgs{LogLevel: logLevel, Tag: tag, Msg: msg, Args: l.castToString(args)})
}

func (l Logger) Debug(tag, msg string, args ...interface{}) {
	l.logMsg("Debug", tag, msg, args)
}

func (l Logger) DebugWithDetails(tag, msg string, args ...interface{}) {
	l.logMsg("DebugWithDetails", tag, msg, args)
}

func (l Logger) Info(tag, msg string, args ...interface{}) {
	l.logMsg("Info", tag, msg, args)
}

func (l Logger) Warn(tag, msg string, args ...interface{}) {
	l.logMsg("Warn", tag, msg, args)
}

func (l Logger) Error(tag, msg string, args ...interface{}) {
	l.logMsg("Error", tag, msg, args)
}

func (l Logger) ErrorWithDetails(tag, msg string, args ...interface{}) {
	l.logMsg("ErrorWithDetails", tag, msg, args)
}

func (l Logger) HandlePanic(tag string) {}

func (l Logger) ToggleForcedDebug() {}

func (l Logger) Flush() error {
	return nil
}

func (l Logger) FlushTimeout(time.Duration) error {
	return nil
}
